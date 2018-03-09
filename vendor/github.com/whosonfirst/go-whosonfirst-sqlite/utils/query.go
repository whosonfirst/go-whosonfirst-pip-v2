package utils

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	"strconv"
	"strings"
)

func ExistentialFlagsToQueryConditions(flag_label string, str_flags string) (string, []interface{}, error) {

	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	for _, str_fl := range strings.Split(str_flags, ",") {

		i, err := strconv.Atoi(str_fl)

		if err != nil {
			return "", args, err
		}

		fl, err := existential.NewKnownUnknownFlag(int64(i))

		if err != nil {
			return "", args, err
		}

		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", flag_label))
		args = append(args, fl.Flag())
	}

	str_conditions := fmt.Sprintf("( %s )", strings.Join(conditions, " OR "))

	return str_conditions, args, nil
}
