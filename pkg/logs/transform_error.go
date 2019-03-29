package logs

// TransformError is logged when the transformer is unable to transform the config event
type TransformError struct {
	Message string `logevent:"message,default=transform-error"`
	Reason  string `logevent:"reason"`
}
