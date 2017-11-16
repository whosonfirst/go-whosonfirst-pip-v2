package flags

type Exclude []string

func (e *Exclude) String() string {
	return strings.Join(*e, "\n")
}

func (e *Exclude) Set(value string) error {
	*e = append(*e, value)
	return nil
}
