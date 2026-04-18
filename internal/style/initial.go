package style

// InitialValue returns the CSS initial value for a given property.
// Only font-family is handled; all other properties return "".
func InitialValue(property string) string {
	switch property {
	case "font-family":
		return "serif"
	default:
		return ""
	}
}
