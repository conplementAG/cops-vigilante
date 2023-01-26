package database_test

import (
	. "github.com/conplementag/cops-vigilante/internal/vigilante/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("InMemory", func() {
	var db Database

	BeforeEach(func() {
		db = NewInMemoryDatabase()
	})

	Describe("With some values set", func() {
		BeforeEach(func() {
			db.Set("a", 123)
			db.Set("b", "any value here")
			db.Set("a", 345)
			db.Set("b", "any value here 2")
		})

		It("should get the values without any problems", func() {
			Expect(db.Get("a")).To(Equal(345))
			Expect(db.Get("b")).To(Equal("any value here 2"))
			Expect(db.Get("c")).To(BeNil())
		})

		It("should return all when requested", func() {
			Expect(len(db.GetAll())).To(Equal(2))
		})

		It("should be able to delete", func() {
			db.Delete("a")
			Expect(db.Get("a")).To(BeNil())
		})
	})
})
