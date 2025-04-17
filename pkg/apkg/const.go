package apkg

var tables = map[string][]string{
	"cards": {
		"id",     // integer primary key
		"nid",    // integer not null, notes.id
		"did",    // integer not null, deck id (available in col table)
		"ord",    // integer not null, ordinal : identifies which of the card templates or cloze deletions it corresponds to
		"mod",    // integer not null, modification time as epoch seconds
		"usn",    // integer not null, update sequence number : used to figure out diffs when syncing.
		"type",   // integer not null, 0=new, 1=learning, 2=review, 3=relearning
		"queue",  // integer not null, -3=user buried(In scheduler 2), -2=sched buried (In scheduler 2), -2=buried(In scheduler 1), -1=suspended, 0=new, 1=learning, 2=review (as for type), 3=in learning, next rev in at least a day after the previous review, 4=preview
		"due",    // integer not null, Due is used differently for different card types: new: the order in which cards are to be studied; starts from 1. learning/relearning: epoch timestamp in seconds review: days since the collection's creation time
		"ivl",    // integer not null, interval (used in SRS algorithm). Negative = seconds, positive = days
		"factor", // integer not null, The ease factor of the card in permille (parts per thousand). If the ease factor is 2500, the card’s interval will be multiplied by 2.5 the next time you press Good.
		"reps",   // integer not null, number of reviews
		"lapses", // integer not null, the number of times the card went from a \"was answered correctly\" to \"was answered incorrectly\" state
		"left",   // integer not null, of the form a*1000+b, with: a the number of reps left today b the number of reps left till graduation
		"odue",   // integer not null, original due: In filtered decks, it's the original due date that the card had before moving to filtered.
		"odid",   // integer not null, original did: only used when the card is currently in filtered deck
		"flags",  // integer not null, an integer. This integer mod 8 represents a \"flag\", which can be see in browser and while reviewing a note. Red 1, Orange 2, Green 3, Blue 4, no flag: 0. This integer divided by 8 represents currently nothing
		"data",   // text not null, currently unused
	},
	"col": {
		"id",     // integer primary key, arbitrary number since there is only one row
		"crt",    // integer not null, timestamp of the creation date in second. It's correct up to the day. For V1 scheduler, the hour corresponds to starting a new day. By default, new day is 4.
		"mod",    // integer not null, last modified in milliseconds
		"scm",    // integer not null, schema mod time: time when \"schema\" was modified.
		"ver",    // integer not null, version
		"dty",    // integer not null, dirty: unused, set to 0
		"usn",    // integer not null, update sequence number: used for finding diffs when syncing.
		"ls",     // integer not null, \"last sync time\"
		"conf",   // text not null, json object containing configuration options that are synced.
		"models", // text not null, json object of json object(s) representing the models (aka Note types)
		"decks",  // text not null, json object of json object(s) representing the deck(s)
		"dconf",  // text not null, json object of json object(s) representing the options group(s) for decks
		"tags",   // text not null, a cache of tags used in the collection (This list is displayed in the browser. Potentially at other place)
	},
	"graves": {
		"usn",  // integer not null, usn should be set to -1
		"oid",  // integer not null, oid is the original id
		"type", // integer not null, type: 0 for a card, 1 for a note and 2 for a deck
	},
	"notes": {
		"id",    // integer primary key, epoch milliseconds of when the note was created
		"guid",  // text not null, globally unique id, almost certainly used for syncing
		"mid",   // integer not null, model id
		"mod",   // integer not null, modification timestamp, epoch seconds
		"usn",   // integer not null, update sequence number: for finding diffs when syncing.
		"tags",  // text not null, space-separated string of tags.
		"flds",  // text not null, the values of the fields in this note. separated by 0x1f (31) character.
		"sfld",  // integer not null, sort field: used for quick sorting and duplicate check. The sort field is an integer so that when users are sorting on a field that contains only numbers, they are sorted in numeric instead of lexical order. Text is stored in this integer field.
		"csum",  // integer not null, field checksum used for duplicate check.
		"flags", // integer not null, unused
		"data",  // text not null, unused
	},
	"revlog": {
		"id",      // integer primary key, epoch-milliseconds timestamp of when you did the review
		"cid",     // integer not null, cards.id
		"usn",     // integer not null, update sequence number: for finding diffs when syncing.
		"ease",    // integer not null, which button you pushed to score your recall.
		"ivl",     // integer not null, interval (i.e. as in the card table)
		"lastIvl", // integer not null, last interval (i.e. the last value of ivl. Note that this value is not necessarily equal to the actual interval between this review and the preceding review)
		"factor",  // integer not null, factor
		"time",    // integer not null, how many milliseconds your review took, up to 60000 (60s)
		"type",    // integer not null, 0=learn, 1=review, 2=relearn, 3=filtered, 4=manual, 5=rescheduled
	},
}
