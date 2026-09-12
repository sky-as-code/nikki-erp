package constants

// NotificationModuleName is the module's name, and the prefix of every schema it owns.
const NotificationModuleName = "notification"

// ChannelName identifies a distribution channel. Channels are registered system-wide in code
// rather than in a table (BR 4): adding one is a new adapter, not a new row.
type ChannelName string

const (
	// ChannelWeb delivers to the browser over the NDJSON stream. It needs no subscription: a
	// notification is available to the web the moment it is committed, whether or not anyone is
	// connected (BR-FS 22).
	ChannelWeb = ChannelName("web")
)

func (this ChannelName) String() string {
	return string(this)
}
