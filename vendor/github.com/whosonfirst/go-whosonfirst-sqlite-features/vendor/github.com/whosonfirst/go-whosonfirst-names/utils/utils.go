package utils

import (
       "strings"
)

func ToRFC5646 (ours string) string {

     theirs := strings.Replace(ours, "_", "-", -1)
     return theirs
}

func FromRFC5646 (theirs string) string {

     ours := strings.Replace(theirs, "-", "_", -1)

     // this needs to account for `-x-FOO` private use subtags that are longer than 8 characters
     return ours
}

