package css

type Declaration struct {
	Property string
	Value    string
}

type Rule struct {
	Selector     string
	Declarations []Declaration
}

type Stylesheet struct {
	Rules []Rule
}
