package wx

type WxValidateRequest struct {
	Timestamp string `Field:"timestamp"`
	Nonce     string `Field:"nonce"`
	Signature string `Field:"msg_signature"`
	Echostr   string `Field:"echostr"`
}

type WxAccessToken struct {
	Token string `json:"access_token"`
	//有效期，单位秒
	Expire int `json:"expires_in"`
}

const (
	MENU_TYPE_CLICK      = "click"
	MENU_TYPE_VIEW       = "view"
	MENU_TYPE_CODE_PUSH  = "scancode_push"
	MENU_TYPE_CODE_WAIT  = "scancode_waitmsg"
	MENU_TYPE_PIC_SYS    = "pic_sysphoto"
	MENU_TYPE_PIC_ALBUM  = "pic_photo_or_album"
	MENU_TYPE_PIC_WEIXIN = "pic_weixin"
	MENU_TYPE_LOCATION   = "location_select"
)

type MenuItem struct {
	Type string      `json:"type"`
	Name string      `json:"name"`
	Key  string      `json:"key"`
	Url  string      `json:"url"`
	Sub  []*MenuItem `json:"sub_button"`
}

type Menu struct {
	Item []*MenuItem `json:"button"`
}

type EventHandle func(msg *WxMessageBody) interface{}

type WxApp struct {
	SimpleController
	AppId string
	Token string

	accessToken *WxAccessToken
	textHandle  EventHandle
	imageHandle EventHandle
	eventHandle EventHandle
}

func (this *WxApp) GetAccessToken() {
	//	"https://api.weixin.qq.com/cgi-bin/token?" + "grant_type=" + "client_credential&appid=APPID&secret=APPSECRET"

}

func (this *WxApp) CreateMenu() {
	//	"https://api.weixin.qq.com/cgi-bin/menu/create?access_token=" + this.accessToken.Token

}

func (this *WxApp) Validate(r *WxRequest) bool {
	//	signatureGen := makeSignature(this.Token, r.Timestamp, r.Nonce)
	//	signatureIn := r.Signature
	//	if signatureGen != signatureIn {
	//		return false
	//	}

	return true
}

func (this *WxApp) SetWxTextHandle(handle EventHandle) {
	this.textHandle = handle

}

func (this *WxApp) SetWxEventHandle(handle EventHandle) {
	this.eventHandle = handle
}

func (this *WxApp) SetWxImageHandle(handle EventHandle) {
	this.imageHandle = handle
}

func (this *WxApp) GetParameType(method string) interface{} {
	if method == "GET" {
		return &WxRequest{}
	} else {
		return &WxMessage{}
	}

}

func (this *WxApp) Get(Context, p interface{}) (interface{}, BingoError) {
	if q, ok := p.(*WxRequest); ok {

		if !this.Validate(q) {
			return ModelView{"index.bingo", "hello"}, nil
			//			return WxResponse{}, nil
		}
		return q.Echostr, nil
	}

	return "hello", nil

}

//正常的访问消息处理
func (this *WxApp) Post(c Context, p interface{}) (interface{}, BingoError) {
	if msg, ok := p.(*WxMessage); ok {
		if this.Validate(&msg.WxRequest) {
			msgbody := msg.data
			var msg interface{}
			messageType := msgbody.MsgType
			if messageType == Event {
				if this.eventHandle != nil {
					msg = this.eventHandle(&msgbody)

				}

			}

			if messageType == Text {

				if this.textHandle != nil {
					msg = this.textHandle(&msgbody)
				}
			} else if messageType == Image {
				if this.imageHandle != nil {
					msg = this.imageHandle(&msgbody)
				}
			}

			if msg == nil {
				textMsg := WxTextMsg{}
				textMsg.Init(&msgbody)
				textMsg.SetBody("暂时还不支持其他的类型")
				msg = textMsg

			}
			return msg, nil
		}

	}
	return WxResponse{}, nil
}

type WxCorpSuit struct {
}
