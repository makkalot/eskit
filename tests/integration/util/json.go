package util

import (
	"encoding/json"
	"github.com/go-test/deep"
	"fmt"

)

func AssertJSONEqual(first, second string) error {

	//ginkgo.GinkgoT().Logf("First json : %s, Second json : %s", first, second)

	firstM := map[string]interface{}{}
	secondM := map[string]interface{}{}

	if err := json.Unmarshal([]byte(first), &firstM); err != nil {
		return fmt.Errorf("first unmarhsall : %v ", err)
	}

	if err := json.Unmarshal([]byte(second), &secondM); err != nil {
		return fmt.Errorf("second unmarhsall : %v ", err)
	}

	if diff := deep.Equal(firstM, secondM); diff != nil {
		return fmt.Errorf("not equal : %v", diff)
	}

	return nil
}
