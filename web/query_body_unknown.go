package web

type msgBody_Unknown struct {
	*msgBody_Base
	Data []byte `json:"data"`
}

func decodeBody_unknown(base *msgBody_Base, raw []byte) (any, error) {
	return &msgBody_Unknown{
		msgBody_Base: base,
		Data:         raw,
	}, nil
}
