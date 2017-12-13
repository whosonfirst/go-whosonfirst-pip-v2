package flags

type ExistentialFlag interface {
	StringFlag() string
	Flag() int64
	IsTrue() bool
	IsFalse() bool
	IsKnown() bool
	MatchesAny(...ExistentialFlag) bool
	MatchesAll(...ExistentialFlag) bool
	String() string
}

type PlacetypeFlag interface {
	MatchesAny(...PlacetypeFlag) bool
	MatchesAll(...PlacetypeFlag) bool
	Placetype() string
	String() string
}
