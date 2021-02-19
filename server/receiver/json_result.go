package receiver

import (
	"encoding/json"
)

// AsrResult /** Wraps the Asr result returned by cerence server

type AsrResult struct {
	Confidences    []int    `json:"confidences"`
	Words          [][]Word `json:"words"`
	Transcriptions []string `json:"transcriptions"`
	FinalResponse  int      `json:"final_response"`
}

type Word struct {
	Confidence string `json:"confidence"`
	Word       string `json:"word"`
}

func (res *AsrResult) GetAtMost(numResults int) *AsrResult {

	itemsToGet := min(numResults, len(res.Confidences))

	if itemsToGet == 0 {
		return &AsrResult{
			Confidences:    []int{},
			Words:          [][]Word{},
			Transcriptions: []string{},
		}
	}

	return &AsrResult{
		Confidences:    res.Confidences[:itemsToGet],
		Words:          res.Words[:itemsToGet],
		Transcriptions: res.Transcriptions[:itemsToGet],
	}

}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (res *AsrResult) ToBytes() ([]byte, error) {
	return json.Marshal(res)
}

func NewAsrResultFrom(data []byte) (*AsrResult, error) {
	var result AsrResult
	err := json.Unmarshal(data, &result)

	return &result, err
}
