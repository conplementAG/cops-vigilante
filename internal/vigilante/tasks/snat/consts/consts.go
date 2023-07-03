package consts

// NodeHealerNamespace we schedule in kube-system because this namespace is always there,
// so we avoid some namespace-related problems
const NodeHealerNamespace = "kube-system"

const NodeHealedAnnotation = "cops-vigilante-snat-node-healed"
const NodeHealerPodNamePrefix = "cops-vigilante-snat-healer-"
