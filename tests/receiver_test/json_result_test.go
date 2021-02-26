package receiver

import "testing"
import "github.com/acepero13/asr-server-cer/server/receiver"

var JSONWithThreeResults = "{\n\"result_type\": \"NVC_ASR_CMD\",\n\"status_code\": 0,\n\"NMAS_PRFX_TRANSACTION_ID\": \"1\",\n\"cadence_regulatable_result\": \"partialRecognition\",\n\"NMAS_PRFX_SESSION_ID\": \"d9e8dc27-46a4-41a8-aad0-7f7ede4c0b55\",\n\"confidences\": [\n545,\n0,\n0\n],\n\"words\": [\n[\n{\n        \"confidence\": \"0.994\",\n        \"word\": \"Ich\\\\*no-space-before\"\n},\n{\n        \"confidence\": \"0.961\",\n        \"word\": \"bin\"\n},\n{\n        \"confidence\": \"0.724\",\n        \"word\": \"hundemüde\"\n},\n{\n        \"confidence\": \"0.957\",\n        \"word\": \"fühle\"\n},\n{\n        \"confidence\": \"0.971\",\n        \"word\": \"mich\"\n},\n{\n        \"confidence\": \"0.957\",\n        \"word\": \"depressiv\"\n},\n{\n        \"confidence\": \"0.972\",\n        \"word\": \"und\"\n},\n{\n        \"confidence\": \"0.56\",\n        \"word\": \"schlafe\"\n},\n{\n        \"confidence\": \"0.984\",\n        \"word\": \"den\"\n},\n{\n        \"confidence\": \"0.99\",\n        \"word\": \"ganzen\"\n},\n{\n        \"confidence\": \"0.986\",\n        \"word\": \"Tag\"\n}\n],\n[\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"Ich\\\\*no-space-before\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"bin\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"Hunde\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"müde\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"fühle\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"mich\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"depressiv\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"und\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"schlafe\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"den\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"ganzen\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"Tag\"\n}\n],\n[\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"Ich\\\\*no-space-before\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"bin\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"hundemüde\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"und\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"fühle\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"mich\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"depressiv\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"und\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"schlafe\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"den\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"ganzen\"\n},\n{\n        \"confidence\": \"0.0\",\n        \"word\": \"Tag\"\n}\n]\n],\n\"transcriptions\": [\n\"Ich bin hundemüde fühle mich depressiv und schlafe den ganzen Tag\",\n\"Ich bin Hunde müde fühle mich depressiv und schlafe den ganzen Tag\",\n\"Ich bin hundemüde und fühle mich depressiv und schlafe den ganzen Tag\"\n],\n\"result_format\": \"rec_text_results\",\n\"final_response\": 0,\n\"prompt\": \"\"\n}"

func TestAsrResult_GetAtMost(t *testing.T) {

}

func TestAsrResult_Parse(t *testing.T) {
	result, _ := receiver.NewAsrResultFrom([]byte(JSONWithThreeResults))

	if len(result.Words) != 3 {
		t.Errorf("Number of word transcriptions do not match")
	}

	if len(result.Words[0]) != 11 {
		t.Errorf("Number of words does not match")
	}

	if result.Words[0][1].Word != "bin" {
		t.Errorf("Wrong word")
	}

	if len(result.Transcriptions) != 3 {
		t.Errorf("Number of transcriptions do not match")
	}
}

func TestAsrResult_ReduceResultToOneElement(t *testing.T) {
	result, _ := receiver.NewAsrResultFrom([]byte(JSONWithThreeResults))

	reduced := result.GetAtMost(1)
	if len(reduced.Words) != 1 {
		t.Errorf("Did not reduce number of words")
	}

	if result.Words[0][1].Word != "bin" {
		t.Errorf("Wrong word")
	}
}

func TestAsrResult_ReduceResultToTwoElements(t *testing.T) {
	result, _ := receiver.NewAsrResultFrom([]byte(JSONWithThreeResults))

	reduced := result.GetAtMost(2)
	if len(reduced.Words) != 2 {
		t.Errorf("Did not reduce number of words")
	}

	if result.Words[0][1].Word != "bin" {
		t.Errorf("Wrong word")
	}
}

func TestAsrResult_RequestMoreThanPossible(t *testing.T) {
	result, _ := receiver.NewAsrResultFrom([]byte(JSONWithThreeResults))

	reduced := result.GetAtMost(4)
	if len(reduced.Words) != 3 {
		t.Errorf("Did not reduce number of words")
	}

	if result.Words[0][1].Word != "bin" {
		t.Errorf("Wrong word")
	}
}

func TestAsrResult_RequestFromEmpty(t *testing.T) {
	result := receiver.AsrResult{
		Confidences:    []int{},
		Transcriptions: []string{},
		Words:          nil,
	}

	reduced := result.GetAtMost(1)
	if len(reduced.Words) != 0 {
		t.Errorf("Did not reduce number of words")
	}

}
