package wx

/**
企业微信加密信息工具类

auth：xiongxiaopeng@qianbaoplus.com
by 钱包行云
2017.10.27

**/
import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
)

/**
40001	签名验证错误
40002	xml解析失败
40003	sha加密生成签名失败
40004	AESKey 非法
40005	corpid 校验错误
40006	AES 加密失败
40007	AES 解密失败
40008	解密后得到的buffer非法
*/

//输出给企业微信服务端的消息
type CorpOutputMessage struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      string
	MsgSignature string
	TimeStamp    string
	Nonce        string
}

//接收到企业微信服务端的消息
type CorpInputputMessage struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string
	AgentID    string
	Encrypt    string
}

type CorpEncrypt struct {
	token          string
	corpid         string
	suiteid        string
	encodingAESKey string
	theAES         AES
}

func (this *CorpEncrypt) Init(token, corpid, suiteid, aeskey string) {
	this.token = token
	this.corpid = corpid
	this.suiteid = suiteid
	if len(aeskey) != 43 { //密码长度不对失败
		panic(aeskey)
	}
	this.encodingAESKey = aeskey
	asekey, _ := base64.URLEncoding.DecodeString(this.encodingAESKey + "=")
	this.theAES = AES{}
	this.theAES.Init(asekey)
}

/**
	 #验证URL
         #@param sMsgSignature: 签名串，对应URL参数的msg_signature
         #@param sTimeStamp: 时间戳，对应URL参数的timestamp
         #@param sNonce: 随机串，对应URL参数的nonce
         #@param sEchoStr: 随机串，对应URL参数的echostr
         #@param sReplyEchoStr: 解密之后的echostr，当return返回0时有效
         #@return：成功0，失败返回对应的错误码

*/
func (this *CorpEncrypt) VerifyURL(msg_signature, timestamp, nonce, echostr string) (int, string) {

	signature := makeSignature(this.token, timestamp, nonce, echostr)
	fmt.Println("signature:" + signature + "|" + msg_signature)
	if signature != msg_signature {
		return 40001, "signature validate failed"

	}
	return 0, this.decrypt(echostr)

}

func (this *CorpEncrypt) DecryptMsg(msg_signature, timestamp, nonce, postdata string) (int, string) {
	input := CorpInputputMessage{}
	err := xml.Unmarshal([]byte(postdata), &input)
	if err != nil {
		return 40002, err.Error()
	}

	return this.DecryptInputMsg(msg_signature, timestamp, nonce, input)
}

func (this *CorpEncrypt) DecryptInputMsg(msg_signature, timestamp, nonce string, postdata CorpInputputMessage) (int, string) {

	signature := makeSignature(this.token, timestamp, nonce, postdata.Encrypt)
	if msg_signature != signature {
		return 40001, "signature validate failed!"
	}

	msg := this.decrypt(postdata.Encrypt)
	if msg == "" {
		return 40005, ""
	}
	return 0, msg
}

func (this *CorpEncrypt) EncryptMsg(replyMsg, nonce, timestamp string) (int, string) {
	msg := CorpOutputMessage{}
	msg.Encrypt = this.encrypt(replyMsg)
	msg.Nonce = nonce
	msg.MsgSignature = makeSignature(this.token, timestamp, nonce, msg.Encrypt)
	msg.TimeStamp = timestamp
	outxml, _ := xml.Marshal(msg)
	return 0, string(outxml)
}

//加密
func (this *CorpEncrypt) encrypt(text string) string {
	var byteGroup bytes.Buffer
	//格式 16位的random字符+文本长度（4位的网络字节序列）+文本+企业id
	randStr := getRandomString(16)
	byteGroup.Write([]byte(randStr))
	byteGroup.Write(numberToBytesOrder(len(text)))
	byteGroup.Write([]byte(text))

	byteGroup.Write([]byte(this.corpid))

	encryptedText := this.theAES.encrypt(byteGroup.Bytes())
	//转base64
	return base64.StdEncoding.EncodeToString(encryptedText)
}

//解密
func (this *CorpEncrypt) decrypt(encryptstr string) string {
	//1、解析base64
	str, _ := base64.StdEncoding.DecodeString(encryptstr)
	sourcebytes := this.theAES.decrypt(str)
	//取字节数组长度
	orderBytes := sourcebytes[16:20]
	msgLength := bytesOrderToNumber(orderBytes)

	//取文本
	text := sourcebytes[20 : 20+msgLength]
	//取企业id
	corpid := sourcebytes[20+msgLength:]
	fmt.Println("the corpid:[" + string(corpid) + "]")
	corp := strings.TrimSpace(string(corpid))
	fmt.Printf("%s,%s:", this.corpid, this.suiteid)
	if this.corpid == corp || this.suiteid == corp {
		fmt.Println("yes!")
		return string(text)

	}

	return ""

}
