package kuberium

type KubeSummaryCmd struct {
	Flags KubeSummaryFlags `embed:""`
}

type KubeSummaryFlags struct {
	Output string `prefix:"k8s.list." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KUBE_SUMMARY_OUTPUT"`
}
