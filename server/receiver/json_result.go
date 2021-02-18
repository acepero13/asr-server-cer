package receiver

import "encoding/json"

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

	return &AsrResult{
		Confidences:    res.Confidences[:numResults],
		Words:          res.Words[:numResults],
		Transcriptions: res.Transcriptions[:numResults],
	}

}

func (res *AsrResult) ToBytes() ([]byte, error) {
	return json.Marshal(res)
}

func NewAsrResultFrom(data []byte) (*AsrResult, error) {
	var result AsrResult
	err := json.Unmarshal(data, &result)

	return &result, err
}
