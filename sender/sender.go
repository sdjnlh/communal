package sender

import (
	"encoding/json"
	"fmt"
	"github.com/sdjnlh/communal/log"
	"go.uber.org/zap"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var response map[string]interface{}

const (
	CpName     = "ll"
	CpPassword = "2dc3c7d576d95f9"
)

func SendCode(mobile string) (string, bool) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	resp, err := http.Post("http://qxt.fungo.cn/Recv_center",
		"application/x-www-form-urlencoded",
		strings.NewReader("CpName="+CpName+"&CpPassword="+CpPassword+"&DesMobile="+mobile+"&Content=【龙灵科技】您的验证码是"+code+",请在十分钟内完成&ExtCode=1234"))
	if err != nil {
		log.Logger.Error("fail send mobile code{}", zap.Any(mobile, err))
		return code, false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)
	respCode, _ := response["code"]
	if "0" != respCode.(string) {
		log.Logger.Error("send rest error{}", zap.Any("error code", respCode))
		return code, false
	}
	return code, true
}
