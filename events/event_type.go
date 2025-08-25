package events

// EventType is a enumeration of all possible event types that are currently supported by Columbo
// If changed in the future, re-run the command `stringer -type=EventType` to regenerate string representations.
type EventType int

const (
	KEventT EventType = iota
	KSimSendSyncT
	KSimProcInEventT
	KHostInstrT
	KHostCallT
	KHostMmioImRespPoWT
	KHostIdOpT
	KHostMmioCRT
	KHostMmioCWT
	KHostAddrSizeOpT
	KHostMmioRT
	KHostMmioWT
	KHostDmaCT
	KHostDmaRT
	KHostDmaWT
	KHostMsiXT
	KHostConfT
	KHostClearIntT
	KHostPostIntT
	KHostPciRWT
	KNicMsixT
	KNicDmaT
	KSetIXT
	KNicDmaIT
	KNicDmaExT
	KNicDmaEnT
	KNicDmaCRT
	KNicDmaCWT
	KNicMmioT
	KNicMmioRT
	KNicMmioWT
	KNicTrxT
	KNicTxT
	KNicRxT
	KNetworKEnqueueT
	KNetworKDequeueT
	KNetworKDropT
)
