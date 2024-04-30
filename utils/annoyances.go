package utils

var (
	Annoyances = []Annoyance{
		{
			Label:     "sponsors",
			FormName:  "sponsors",
			Color:     "checkbox-success",
			AriaLabel: "Sponsors",
			ShortKey:  "sponsors",
		},
		{
			Label:     "unpaid / self-promotion",
			FormName:  "pro",
			Color:     "checkbox-warning",
			AriaLabel: "Unpaid / Self-promotion",
			ShortKey:  "self_promotion",
		},
		{
			Label:     "intermissions / intros",
			FormName:  "intermis",
			Color:     "checkbox-primary",
			AriaLabel: "Intermissions / Intros",
			ShortKey:  "intermissions",
		},
		{
			Label:     "interaction reminders",
			FormName:  "reminder",
			Color:     "checkbox-secondary",
			AriaLabel: "Interaction Reminders",
			ShortKey:  "reminders",
		},
		{
			Label:     "credits / endcards",
			FormName:  "credits",
			Color:     "checkbox-accent",
			AriaLabel: "Credits / Endcards",
			ShortKey:  "credits",
		},
		{
			Label:     "recaps",
			FormName:  "recaps",
			Color:     "checkbox-info",
			AriaLabel: "Recaps",
			ShortKey:  "recaps",
		},
	}
)

type Annoyance struct {
	Label     string // used for UI label
	FormName  string // name property of the element
	AriaLabel string // aria-label property of the element
	ShortKey  string // short key is the key in the localStorage
	Color     string // color of the element (class name)
}
