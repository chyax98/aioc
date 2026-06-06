package skill

import _ "embed"

//go:embed aioc-cli/SKILL.md
var AIOCCLI string

const Name = "aioc-cli"
const FileName = "SKILL.md"
