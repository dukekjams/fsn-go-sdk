package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/FusionFoundation/fsn-go-sdk/efsn/cmd/utils"
	"github.com/FusionFoundation/fsn-go-sdk/efsn/common"
	"github.com/FusionFoundation/fsn-go-sdk/efsn/core/types"
	"github.com/FusionFoundation/fsn-go-sdk/efsn/rlp"
)

func DecodeLogData(logData []byte) (interface{}, error) {
	logMap := make(map[string]interface{})
	if err := json.Unmarshal(logData, &logMap); err != nil {
		return nil, fmt.Errorf("json unmarshal err: %v", err)
	}
	if basestr, ok := logMap["Base"].(string); ok {
		if data, err := base64.StdEncoding.DecodeString(basestr); err == nil {
			buyTicketParam := common.BuyTicketParam{}
			if err = rlp.DecodeBytes(data, &buyTicketParam); err == nil {
				logMap["StartTime"] = buyTicketParam.Start
				logMap["ExpireTime"] = buyTicketParam.End
			}
		}
	}
	delete(logMap, "Base")
	return logMap, nil
}

func DecodeFSNLogData(funcType common.FSNCallFunc, logData []byte) (interface{}, error) {
	logMap := make(map[string]interface{})
	if err := json.Unmarshal(logData, &logMap); err != nil {
		return nil, fmt.Errorf("json unmarshal err: %v", err)
	}
	if _, hasBase := logMap["Base"]; !hasBase {
		return logMap, nil
	}
	switch funcType {
	case common.GenNotationFunc:
		delete(logMap, "Base")

	case common.BuyTicketFunc:
		basestr, ok := logMap["Base"].(string)
		if !ok {
			break
		}
		data, err := base64.StdEncoding.DecodeString(basestr)
		if err != nil {
			return nil, fmt.Errorf("base64 decode err: %v", err)
		}
		buyTicketParam := common.BuyTicketParam{}
		if err = rlp.DecodeBytes(data, &buyTicketParam); err == nil {
			delete(logMap, "Base")
			logMap["StartTime"] = buyTicketParam.Start
			logMap["ExpireTime"] = buyTicketParam.End
		}
	}
	return logMap, nil
}

func DecodeTxInput(input []byte) (interface{}, error) {
	res, err := common.DecodeTxInput(input)
	if err == nil {
		return res, err
	}
	fsnCall, ok := res.(common.FSNCallParam)
	if !ok {
		return res, err
	}
	switch fsnCall.Func {
	case common.ReportIllegalFunc:
		h1, h2, err := DecodeReport(fsnCall.Data)
		if err != nil {
			return nil, fmt.Errorf("DecodeReport err %v", err)
		}
		reportContent := &struct {
			Header1 *types.Header
			Header2 *types.Header
		}{
			Header1: h1,
			Header2: h2,
		}
		fsnCall.Data = nil
		return common.DecodeFsnCallParam(&fsnCall, reportContent)
	}
	return nil, fmt.Errorf("Unknown FuncType %v", fsnCall.Func)
}

func DecodeReport(report []byte) (*types.Header, *types.Header, error) {
	if len(report) < 4 {
		return nil, nil, fmt.Errorf("wrong report length")
	}
	data1len := common.BytesToInt(report[:4])
	if len(report) < 4+data1len {
		return nil, nil, fmt.Errorf("wrong report length")
	}
	data1 := report[4 : data1len+4]
	data2 := report[data1len+4:]

	if bytes.Compare(data1, data2) >= 0 {
		return nil, nil, fmt.Errorf("wrong report sequence")
	}

	header1 := &types.Header{}
	header2 := &types.Header{}

	if err := rlp.DecodeBytes(data1, header1); err != nil {
		return nil, nil, fmt.Errorf("can not decode header1, err=%v", err)
	}
	if err := rlp.DecodeBytes(data2, header2); err != nil {
		return nil, nil, fmt.Errorf("can not decode header2, err=%v", err)
	}
	return header1, header2, nil
}

// MustPrintJSON prints the JSON encoding of the given object and
// exits the program with an error message when the marshaling fails.
func MustPrintJSON(jsonObject interface{}) {
	str, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		utils.Fatalf("Failed to marshal JSON object: %v", err)
	}
	fmt.Println(string(str))
}
