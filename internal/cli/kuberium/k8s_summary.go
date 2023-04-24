package kuberium

type K8sSummaryCmd struct {
	Flags K8sSummaryFlags `embed:""`
}

type K8sSummaryFlags struct {
	Output string `prefix:"k8s.list." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_K8S_SUMMARY_OUTPUT"`
}
