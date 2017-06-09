package rtm

const (
	ERROR_CODE_APPLICATION    = 0
	ERROR_CODE_TRANSPORT      = 1
	ERROR_CODE_PDU            = 2
	ERROR_CODE_INVALID_JSON   = 3
	ERROR_CODE_AUTHENTICATION = 4
)

type RTMError struct {
	Code   int
	Reason error
}

func (re RTMError) Error() string {
	if re.Reason != nil {
		return re.Reason.Error()
	}
	return "Unknow error"
}
