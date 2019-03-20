package logs

// UnsupportedResource is logged when the transformer receives an unsupported AWS resource
type UnsupportedResource struct {
	Message  string `logevent:"message,default=unsupported-resource"`
	Resource string `logevent:"resource"`
}
