package role

type Role int

const (
	Physician Role = iota // 0 — врач, создатель критериев
	Moderator             // 1 — модератор справочника
)
