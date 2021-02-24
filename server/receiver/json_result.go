package receiver

import (
	"encoding/json"
)

// AsrResult  Wraps the Asr result returned by cerence server
type AsrResult struct {
	Confidences    []int    `json:"confidences"`
	Words          [][]Word `json:"words"`
	Transcriptions []string `json:"transcriptions"`
	FinalResponse  int      `json:"final_response"`
}

//Word Representation of a single word of the recognition
type Word struct {
	Confidence string `json:"confidence"`
	Word       string `json:"word"`
}

//GetAtMost Discard extra results keeping only at most numResults results.
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

//ToBytes Converts an AsrResult to its byte array representation
func (res *AsrResult) ToBytes() ([]byte, error) {
	return json.Marshal(res)
}

//NewAsrResultFrom Creates a new asr result from a byte array representation
func NewAsrResultFrom(data []byte) (*AsrResult, error) {
	var result AsrResult
	err := json.Unmarshal(data, &result)

	return &result, err
}
