package constants

const SalesModuleName = "sales"

// SalesRouteV1 is the REST route group every Sales resource engine hangs off.
//
// It matches the schema prefix rather than merely resembling it: the engine derives a resource's
// path segment from its schema name, so "/v1/sales" plus schema "sales_channels" is the URL a
// client calls and "sales_channels" is the IAM resource code asserted against it. Those three
// strings are one string in three places.
const SalesRouteV1 = "/v1/sales"
