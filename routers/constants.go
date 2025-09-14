package routers

var (
	// common header keys
	HEADER_ROUTE_INDEX   = "Flexy-Route-Index"
	HEADER_RESPONSE_TYPE = "Flexy-Response-Handler-Type"
	// rewrite
	HEADER_REWRITE_TO = "Flexy-Redirect-To"
	// file
	HEADER_FILE_PATH = "Flexy-File-Path"
)

var HEADER_KEY_TO_LOG_KEY = map[string]string{
	HEADER_ROUTE_INDEX:   "route_index",
	HEADER_RESPONSE_TYPE: "response_type",
	HEADER_REWRITE_TO:    "rewrite_to",
	HEADER_FILE_PATH:     "file_path",
}
