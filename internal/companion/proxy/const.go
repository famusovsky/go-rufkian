package proxy

const (
	woerterURL     = "https://www.woerter.net/?w="
	styleElementID = "text/css"
)

var (
	classesToReturn = []string{
		"rAbschnitt ", "rInfo",
	}

	elementsWithTypeToExclude = []string{
		"nav", "svg",
	}

	elementsWithClassToExclude = []string{
		"rKnpf", "rRechts",
	}

	elementsWithTypeToPop = []string{
		"section",
	}

	attributesToSave = []string{
		"style",
	}

	languagesToSave = []string{
		"ru", "en",

		// just for fun
		"uk", "be",
		"he", "ar",
	}
)
