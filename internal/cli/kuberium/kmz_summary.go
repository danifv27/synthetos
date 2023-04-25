package kuberium

type KmzSummaryCmd struct {
	Flags KubeSummaryFlags `embed:""`
}

type KmzSummaryFlags struct {
	Output string `prefix:"k8s.list." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KMZ_SUMMARY_OUTPUT"`
}
