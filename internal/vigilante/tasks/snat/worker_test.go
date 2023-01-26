package snat_test

import (
	"errors"
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock/testing"
	"github.com/conplementag/cops-vigilante/internal/vigilante/database"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/consts"
	. "github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("SNAT Worker", func() {
	var kubernetesServiceMock *KubernetesServiceMock
	var metricsRecorderMock *SnatMetricsRecorderMock
	var fakeClock *testing.FakeClock
	var task tasks.Task

	BeforeEach(func() {
		fakeClock = &testing.FakeClock{CurrentTime: time.Now()}

		kubernetesServiceMock = NewKubernetesServiceMock(fakeClock)
		metricsRecorderMock = &SnatMetricsRecorderMock{}
		metricsRecorderMock.On("IncHealAttemptsCounter", mock.Anything)
		metricsRecorderMock.On("IncNumberOfNotReadyNodesSeen", mock.Anything)

		task = snat.NewSnatTask(kubernetesServiceMock, database.NewInMemoryDatabase(), metricsRecorderMock, fakeClock)
	})

	When("running on a fresh cluster", func() {
		BeforeEach(func() {
			kubernetesServiceMock.TestData_Nodes = GenerateNodesSimilarToInitialNonHealedCluster()
			task.Run()
		})

		It("should schedule the heal pod on every windows node", func() {
			Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksWindowsNode1]).ToNot(BeNil())
			Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksWindowsNode2]).ToNot(BeNil())

			metricsRecorderMock.AssertNumberOfCalls(GinkgoT(), "IncHealAttemptsCounter", 2)
		})

		It("should ignore the linux nodes", func() {
			Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksLinuxNode1]).To(BeNil())
			Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksLinuxNode2]).To(BeNil())
			Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksLinuxNode3]).To(BeNil())
		})

		It("should ignore the non-ready windows nodes", func() {
			Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksWindowsNode3NeverReady]).To(BeNil())
			metricsRecorderMock.AssertNumberOfCalls(GinkgoT(), "IncNumberOfNotReadyNodesSeen", 1)
		})

		It("should mark no node as healed yet", func() {
			for _, node := range kubernetesServiceMock.TestData_Nodes {
				_, keyFound := node.Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
			}
		})

		Context("running for the second time before all the nodes are considered healed", func() {
			BeforeEach(func() {
				// at this point, before running again, we should make sure no pods were deleted yet, as we started in a "fresh" cluster
				Expect(kubernetesServiceMock.TestData_DeletedPods).To(BeEmpty())

				fakeClock.PassTime(5 * time.Minute)
				task.Run()
			})

			It("should remove the existing pod before scheduling a new one", func() {
				Expect(kubernetesServiceMock.TestData_DeletedPods[consts.NodeHealerPodNamePrefix+AksWindowsNode1]).ToNot(BeNil())
				Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksWindowsNode1].CreationTimestamp).
					To(Equal(apimachinerymetav1.NewTime(fakeClock.Now())))

				Expect(kubernetesServiceMock.TestData_DeletedPods[consts.NodeHealerPodNamePrefix+AksWindowsNode2]).ToNot(BeNil())
				Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+AksWindowsNode2].CreationTimestamp).
					To(Equal(apimachinerymetav1.NewTime(fakeClock.Now())))
			})

			It("should record the metrics correctly", func() {
				// each task.Run() should schedule two new healing pods, and the not ready node should be seen again
				metricsRecorderMock.AssertNumberOfCalls(GinkgoT(), "IncHealAttemptsCounter", 4)
				metricsRecorderMock.AssertNumberOfCalls(GinkgoT(), "IncNumberOfNotReadyNodesSeen", 2)
			})
		})

		Context("healing time is passed without errors", func() {
			BeforeEach(func() {
				// set the time to be after the considered healing period
				fakeClock.PassTime(snat.DefaultHealingDurationPerNode + time.Minute)
				task.Run()
			})

			It("should mark the node(s) as healed", func() {
				Expect(services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode1).Annotations[consts.NodeHealedAnnotation]).
					To(Equal("true"))
				Expect(services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode2).Annotations[consts.NodeHealedAnnotation]).
					To(Equal("true"))

				_, keyFound := services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode3NeverReady).Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
				_, keyFound = services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksLinuxNode1).Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
				_, keyFound = services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksLinuxNode2).Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
				_, keyFound = services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksLinuxNode3).Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
			})

			It("should remove the pod from the healed node(s)", func() {
				Expect(kubernetesServiceMock.TestData_DeletedPods[consts.NodeHealerPodNamePrefix+AksWindowsNode1]).ToNot(BeNil())
				Expect(kubernetesServiceMock.TestData_DeletedPods[consts.NodeHealerPodNamePrefix+AksWindowsNode2]).ToNot(BeNil())
			})
		})

		Context("healing time is passed with some errors", func() {
			BeforeEach(func() {
				// run for a couple of times without errors
				for i := 0; i < 10; i++ {
					fakeClock.PassTime(30 * time.Second)
					task.Run()
				}

				// then for a couple of times with errors when scheduling the pods
				kubernetesServiceMock.Error_CreatePod = errors.New("some issue occurred")
				for i := 0; i < 5; i++ {
					fakeClock.PassTime(30 * time.Second)
					task.Run()
				}

				// set the time to be after the considered healing period and run
				fakeClock.PassTime(snat.DefaultHealingDurationPerNode + time.Minute)
				task.Run()
			})

			It("should still mark the node as healed if the error threshold is not reached", func() {
				Expect(services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode1).Annotations[consts.NodeHealedAnnotation]).
					To(Equal("true"))
				Expect(services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode2).Annotations[consts.NodeHealedAnnotation]).
					To(Equal("true"))
			})
		})

		Context("healing time is passed with a lot of errors", func() {
			BeforeEach(func() {
				// run for a couple of times without errors
				for i := 0; i < 5; i++ {
					fakeClock.PassTime(30 * time.Second)
					task.Run()
				}

				// then for a lot of times with many pod creation errors
				kubernetesServiceMock.Error_CreatePod = errors.New("some issue occurred")
				for i := 0; i < 30; i++ {
					fakeClock.PassTime(30 * time.Second)
					task.Run()
				}

				// set the time to be after the considered healing period and run
				fakeClock.PassTime(snat.DefaultHealingDurationPerNode + time.Minute)
				task.Run()
			})

			It("should mark no node as healed", func() {
				_, keyFound := services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode1).Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
				_, keyFound = services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, AksWindowsNode2).Annotations[consts.NodeHealedAnnotation]
				Expect(keyFound).To(BeFalse())
			})
		})
	})

	When("running on a fully healed cluster", func() {
		BeforeEach(func() {
			kubernetesServiceMock.TestData_Nodes = GenerateNodesLikeInHealedCluster()
			task.Run()
		})

		It("should schedule no pods for healing", func() {
			Expect(kubernetesServiceMock.TestData_Pods).To(BeEmpty())
		})

		It("should not delete any pods", func() {
			Expect(kubernetesServiceMock.TestData_DeletedPods).To(BeEmpty())
		})

		Context("a new node is added", func() {
			var newNodeName string
			var timeWhenNewNodeWasAdded time.Time

			BeforeEach(func() {
				newNodeName = "akswin-newnode"
				kubernetesServiceMock.TestData_Nodes = append(kubernetesServiceMock.TestData_Nodes,
					GenerateNode(newNodeName, true, false, false, true))

				fakeClock.PassTime(2 * time.Minute)
				timeWhenNewNodeWasAdded = fakeClock.CurrentTime
				task.Run()
			})

			It("should schedule a healing pod", func() {
				Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+newNodeName]).ToNot(BeNil())
				Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+newNodeName].CreationTimestamp).
					To(Equal(apimachinerymetav1.NewTime(timeWhenNewNodeWasAdded)))
			})

			Context("then the new node becomes unready", func() {
				BeforeEach(func() {
					UpdateNodeCondition(services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, newNodeName), false)

					fakeClock.PassTime(2 * time.Minute)
					task.Run()
				})

				It("should not attempt to schedule any new healing pods", func() {
					Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+newNodeName].CreationTimestamp).
						To(Equal(apimachinerymetav1.NewTime(timeWhenNewNodeWasAdded)))
				})

				It("should not try to remove the existing pod", func() {
					Expect(kubernetesServiceMock.TestData_DeletedPods).To(BeEmpty())
				})

				Context("and the same node becomes ready again", func() {
					var timeWhenNodeBecameReadyAgain time.Time

					BeforeEach(func() {
						UpdateNodeCondition(services.FindNodeByName(kubernetesServiceMock.TestData_Nodes, newNodeName), true)

						fakeClock.PassTime(2 * time.Minute)
						timeWhenNodeBecameReadyAgain = fakeClock.CurrentTime
						task.Run()
					})

					It("should remove the old pod", func() {
						Expect(kubernetesServiceMock.TestData_DeletedPods[consts.NodeHealerPodNamePrefix+newNodeName]).ToNot(BeNil())
					})

					It("should run the normal healing process again (duration restarted), all the way until the end", func() {
						By("scheduling a new healing pod")
						Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+newNodeName].CreationTimestamp).
							To(Equal(apimachinerymetav1.NewTime(timeWhenNodeBecameReadyAgain)))

						By("still scheduling a new healing pod if the duration did not pass yet")
						fakeClock.PassTime(snat.DefaultHealingDurationPerNode - time.Minute)
						task.Run()
						Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+newNodeName].CreationTimestamp).
							To(Equal(apimachinerymetav1.NewTime(fakeClock.CurrentTime)))

						By("removing the healer pod after the heal duration has passed")
						fakeClock.PassTime(2 * time.Minute)
						task.Run()
						Expect(kubernetesServiceMock.TestData_Pods[consts.NodeHealerPodNamePrefix+newNodeName]).To(BeNil())
						Expect(kubernetesServiceMock.TestData_DeletedPods[consts.NodeHealerPodNamePrefix+newNodeName]).ToNot(BeNil())
					})
				})
			})
		})
	})
})
