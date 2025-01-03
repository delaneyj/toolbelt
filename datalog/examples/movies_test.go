package examples

import (
	"testing"

	"github.com/delaneyj/toolbelt/datalog"
	"github.com/stretchr/testify/assert"
)

func TestValidMovieId(t *testing.T) {
	actual := datalog.MatchPattern(
		datalog.Pattern{"?movieId", "movie/director", "?directorId"},
		datalog.Triple{"200", "movie/director", "100"},
		datalog.State{"?movieId": "200"},
	)

	expected := datalog.State{"?movieId": "200", "?directorId": "100"}
	assert.Equal(t, expected, actual)
}

func TestInvalidMovieId(t *testing.T) {
	actual := datalog.MatchPattern(
		datalog.Pattern{"?movieId", "movie/director", "?directorId"},
		datalog.Triple{"200", "movie/director", "100"},
		datalog.State{"?movieId": "202"},
	)

	assert.Nil(t, actual)
}

func TestQuerySingle(t *testing.T) {
	db := datalog.CreateDB(Movies...)
	actual := db.QuerySingle(
		datalog.State{},
		datalog.Pattern{"?movieId", "movie/year", "1987"},
	)

	expected := []datalog.State{
		{"?movieId": "202"},
		{"?movieId": "203"},
		{"?movieId": "204"},
	}
	assert.Equal(t, expected, actual)
}

func TestQueryWhere(t *testing.T) {
	db := datalog.CreateDB(Movies...)
	actual := db.QueryWhere(
		datalog.Pattern{"?movieId", "movie/title", "The Terminator"},
		datalog.Pattern{"?movieId", "movie/director", "?directorId"},
		datalog.Pattern{"?directorId", "person/name", "?directorName"},
	)

	expected := []datalog.State{
		{"?movieId": "200", "?directorId": "100", "?directorName": "James Cameron"},
	}
	assert.Equal(t, expected, actual)
}

func TestQueryWhoDirectedTerminator(t *testing.T) {
	db := datalog.CreateDB(Movies...)
	actual := db.Query(
		[]string{"?directorName"},
		datalog.Pattern{"?movieId", "movie/title", "The Terminator"},
		datalog.Pattern{"?movieId", "movie/director", "?directorId"},
		datalog.Pattern{"?directorId", "person/name", "?directorName"},
	)

	expected := [][]string{
		{"James Cameron"},
	}
	assert.Equal(t, expected, actual)
}

func TestQueryWhenAlienWasReleased(t *testing.T) {
	db := datalog.CreateDB(Movies...)
	actual := db.Query(
		[]string{"?attr", "?value"},
		datalog.Pattern{"200", "?attr", "?value"},
	)

	expected := [][]string{
		{"movie/title", "The Terminator"},
		{"movie/year", "1984"},
		{"movie/director", "100"},
		{"movie/cast", "101"},
		{"movie/cast", "102"},
		{"movie/cast", "103"},
		{"movie/sequel", "207"},
	}
	assert.Equal(t, expected, actual)
}

func TestQueryWhatDoIKnowAboutEntityWithID200(t *testing.T) {
	db := datalog.CreateDB(Movies...)
	actual := db.Query(
		[]string{"?predicate", "?object"},
		datalog.Pattern{"200", "?predicate", "?object"},
	)

	expected := [][]string{
		{"movie/title", "The Terminator"},
		{"movie/year", "1984"},
		{"movie/director", "100"},
		{"movie/cast", "101"},
		{"movie/cast", "102"},
		{"movie/cast", "103"},
		{"movie/sequel", "207"},
	}
	assert.Equal(t, expected, actual)
}

func TestQueryWhichDirectorsForArnoldForWhichMovies(t *testing.T) {
	db := datalog.CreateDB(Movies...)
	actual := db.Query(
		[]string{"?directorName", "?movieTitle"},
		datalog.Pattern{"?arnoldId", "person/name", "Arnold Schwarzenegger"},
		datalog.Pattern{"?movieId", "movie/cast", "?arnoldId"},
		datalog.Pattern{"?movieId", "movie/title", "?movieTitle"},
		datalog.Pattern{"?movieId", "movie/director", "?directorId"},
		datalog.Pattern{"?directorId", "person/name", "?directorName"},
	)

	expected := [][]string{
		{"James Cameron", "The Terminator"},
		{"John McTiernan", "Predator"},
		{"Mark L. Lester", "Commando"},
		{"James Cameron", "Terminator 2: Judgment Day"},
		{"Jonathan Mostow", "Terminator 3: Rise of the Machines"},
	}
	assert.Equal(t, expected, actual)
}

var Movies = []datalog.Triple{
	{"100", "person/name", "James Cameron"},
	{"100", "person/born", "1954-08-16T00:00:00Z"},
	{"101", "person/name", "Arnold Schwarzenegger"},
	{"101", "person/born", "1947-07-30T00:00:00Z"},
	{"102", "person/name", "Linda Hamilton"},
	{"102", "person/born", "1956-09-26T00:00:00Z"},
	{"103", "person/name", "Michael Biehn"},
	{"103", "person/born", "1956-07-31T00:00:00Z"},
	{"104", "person/name", "Ted Kotcheff"},
	{"104", "person/born", "1931-04-07T00:00:00Z"},
	{"105", "person/name", "Sylvester Stallone"},
	{"105", "person/born", "1946-07-06T00:00:00Z"},
	{"106", "person/name", "Richard Crenna"},
	{"106", "person/born", "1926-11-30T00:00:00Z"},
	{"106", "person/death", "2003-01-17T00:00:00Z"},
	{"107", "person/name", "Brian Dennehy"},
	{"107", "person/born", "1938-07-09T00:00:00Z"},
	{"108", "person/name", "John McTiernan"},
	{"108", "person/born", "1951-01-08T00:00:00Z"},
	{"109", "person/name", "Elpidia Carrillo"},
	{"109", "person/born", "1961-08-16T00:00:00Z"},
	{"110", "person/name", "Carl Weathers"},
	{"110", "person/born", "1948-01-14T00:00:00Z"},
	{"111", "person/name", "Richard Donner"},
	{"111", "person/born", "1930-04-24T00:00:00Z"},
	{"112", "person/name", "Mel Gibson"},
	{"112", "person/born", "1956-01-03T00:00:00Z"},
	{"113", "person/name", "Danny Glover"},
	{"113", "person/born", "1946-07-22T00:00:00Z"},
	{"114", "person/name", "Gary Busey"},
	{"114", "person/born", "1944-07-29T00:00:00Z"},
	{"115", "person/name", "Paul Verhoeven"},
	{"115", "person/born", "1938-07-18T00:00:00Z"},
	{"116", "person/name", "Peter Weller"},
	{"116", "person/born", "1947-06-24T00:00:00Z"},
	{"117", "person/name", "Nancy Allen"},
	{"117", "person/born", "1950-06-24T00:00:00Z"},
	{"118", "person/name", "Ronny Cox"},
	{"118", "person/born", "1938-07-23T00:00:00Z"},
	{"119", "person/name", "Mark L. Lester"},
	{"119", "person/born", "1946-11-26T00:00:00Z"},
	{"120", "person/name", "Rae Dawn Chong"},
	{"120", "person/born", "1961-02-28T00:00:00Z"},
	{"121", "person/name", "Alyssa Milano"},
	{"121", "person/born", "1972-12-19T00:00:00Z"},
	{"122", "person/name", "Bruce Willis"},
	{"122", "person/born", "1955-03-19T00:00:00Z"},
	{"123", "person/name", "Alan Rickman"},
	{"123", "person/born", "1946-02-21T00:00:00Z"},
	{"124", "person/name", "Alexander Godunov"},
	{"124", "person/born", "1949-11-28T00:00:00Z"},
	{"124", "person/death", "1995-05-18T00:00:00Z"},
	{"125", "person/name", "Robert Patrick"},
	{"125", "person/born", "1958-11-05T00:00:00Z"},
	{"126", "person/name", "Edward Furlong"},
	{"126", "person/born", "1977-08-02T00:00:00Z"},
	{"127", "person/name", "Jonathan Mostow"},
	{"127", "person/born", "1961-11-28T00:00:00Z"},
	{"128", "person/name", "Nick Stahl"},
	{"128", "person/born", "1979-12-05T00:00:00Z"},
	{"129", "person/name", "Claire Danes"},
	{"129", "person/born", "1979-04-12T00:00:00Z"},
	{"130", "person/name", "George P. Cosmatos"},
	{"130", "person/born", "1941-01-04T00:00:00Z"},
	{"130", "person/death", "2005-04-19T00:00:00Z"},
	{"131", "person/name", "Charles Napier"},
	{"131", "person/born", "1936-04-12T00:00:00Z"},
	{"131", "person/death", "2011-10-05T00:00:00Z"},
	{"132", "person/name", "Peter MacDonald"},
	{"133", "person/name", "Marc de Jonge"},
	{"133", "person/born", "1949-02-16T00:00:00Z"},
	{"133", "person/death", "1996-06-06T00:00:00Z"},
	{"134", "person/name", "Stephen Hopkins"},
	{"135", "person/name", "Ruben Blades"},
	{"135", "person/born", "1948-07-16T00:00:00Z"},
	{"136", "person/name", "Joe Pesci"},
	{"136", "person/born", "1943-02-09T00:00:00Z"},
	{"137", "person/name", "Ridley Scott"},
	{"137", "person/born", "1937-11-30T00:00:00Z"},
	{"138", "person/name", "Tom Skerritt"},
	{"138", "person/born", "1933-08-25T00:00:00Z"},
	{"139", "person/name", "Sigourney Weaver"},
	{"139", "person/born", "1949-10-08T00:00:00Z"},
	{"140", "person/name", "Veronica Cartwright"},
	{"140", "person/born", "1949-04-20T00:00:00Z"},
	{"141", "person/name", "Carrie Henn"},
	{"142", "person/name", "George Miller"},
	{"142", "person/born", "1945-03-03T00:00:00Z"},
	{"143", "person/name", "Steve Bisley"},
	{"143", "person/born", "1951-12-26T00:00:00Z"},
	{"144", "person/name", "Joanne Samuel"},
	{"145", "person/name", "Michael Preston"},
	{"145", "person/born", "1938-05-14T00:00:00Z"},
	{"146", "person/name", "Bruce Spence"},
	{"146", "person/born", "1945-09-17T00:00:00Z"},
	{"147", "person/name", "George Ogilvie"},
	{"147", "person/born", "1931-03-05T00:00:00Z"},
	{"148", "person/name", "Tina Turner"},
	{"148", "person/born", "1939-11-26T00:00:00Z"},
	{"149", "person/name", "Sophie Marceau"},
	{"149", "person/born", "1966-11-17T00:00:00Z"},
	{"200", "movie/title", "The Terminator"},
	{"200", "movie/year", "1984"},
	{"200", "movie/director", "100"},
	{"200", "movie/cast", "101"},
	{"200", "movie/cast", "102"},
	{"200", "movie/cast", "103"},
	{"200", "movie/sequel", "207"},
	{"201", "movie/title", "First Blood"},
	{"201", "movie/year", "1982"},
	{"201", "movie/director", "104"},
	{"201", "movie/cast", "105"},
	{"201", "movie/cast", "106"},
	{"201", "movie/cast", "107"},
	{"201", "movie/sequel", "209"},
	{"202", "movie/title", "Predator"},
	{"202", "movie/year", "1987"},
	{"202", "movie/director", "108"},
	{"202", "movie/cast", "101"},
	{"202", "movie/cast", "109"},
	{"202", "movie/cast", "110"},
	{"202", "movie/sequel", "211"},
	{"203", "movie/title", "Lethal Weapon"},
	{"203", "movie/year", "1987"},
	{"203", "movie/director", "111"},
	{"203", "movie/cast", "112"},
	{"203", "movie/cast", "113"},
	{"203", "movie/cast", "114"},
	{"203", "movie/sequel", "212"},
	{"204", "movie/title", "RoboCop"},
	{"204", "movie/year", "1987"},
	{"204", "movie/director", "115"},
	{"204", "movie/cast", "116"},
	{"204", "movie/cast", "117"},
	{"204", "movie/cast", "118"},
	{"205", "movie/title", "Commando"},
	{"205", "movie/year", "1985"},
	{"205", "movie/director", "119"},
	{"205", "movie/cast", "101"},
	{"205", "movie/cast", "120"},
	{"205", "movie/cast", "121"},
	{"205", "trivia", "In 1986, a sequel was written with an eye to having\n  John McTiernan direct. Schwarzenegger wasn't interested in reprising\n  the role. The script was then reworked with a new central character,\n  eventually played by Bruce Willis, and became Die Hard"},
	{"206", "movie/title", "Die Hard"},
	{"206", "movie/year", "1988"},
	{"206", "movie/director", "108"},
	{"206", "movie/cast", "122"},
	{"206", "movie/cast", "123"},
	{"206", "movie/cast", "124"},
	{"207", "movie/title", "Terminator 2: Judgment Day"},
	{"207", "movie/year", "1991"},
	{"207", "movie/director", "100"},
	{"207", "movie/cast", "101"},
	{"207", "movie/cast", "102"},
	{"207", "movie/cast", "125"},
	{"207", "movie/cast", "126"},
	{"207", "movie/sequel", "208"},
	{"208", "movie/title", "Terminator 3: Rise of the Machines"},
	{"208", "movie/year", "2003"},
	{"208", "movie/director", "127"},
	{"208", "movie/cast", "101"},
	{"208", "movie/cast", "128"},
	{"208", "movie/cast", "129"},
	{"209", "movie/title", "Rambo: First Blood Part II"},
	{"209", "movie/year", "1985"},
	{"209", "movie/director", "130"},
	{"209", "movie/cast", "105"},
	{"209", "movie/cast", "106"},
	{"209", "movie/cast", "131"},
	{"209", "movie/sequel", "210"},
	{"210", "movie/title", "Rambo III"},
	{"210", "movie/year", "1988"},
	{"210", "movie/director", "132"},
	{"210", "movie/cast", "105"},
	{"210", "movie/cast", "106"},
	{"210", "movie/cast", "133"},
	{"211", "movie/title", "Predator 2"},
	{"211", "movie/year", "1990"},
	{"211", "movie/director", "134"},
	{"211", "movie/cast", "113"},
	{"211", "movie/cast", "114"},
	{"211", "movie/cast", "135"},
	{"212", "movie/title", "Lethal Weapon 2"},
	{"212", "movie/year", "1989"},
	{"212", "movie/director", "111"},
	{"212", "movie/cast", "112"},
	{"212", "movie/cast", "113"},
	{"212", "movie/cast", "136"},
	{"212", "movie/sequel", "213"},
	{"213", "movie/title", "Lethal Weapon 3"},
	{"213", "movie/year", "1992"},
	{"213", "movie/director", "111"},
	{"213", "movie/cast", "112"},
	{"213", "movie/cast", "113"},
	{"213", "movie/cast", "136"},
	{"214", "movie/title", "Alien"},
	{"214", "movie/year", "1979"},
	{"214", "movie/director", "137"},
	{"214", "movie/cast", "138"},
	{"214", "movie/cast", "139"},
	{"214", "movie/cast", "140"},
	{"214", "movie/sequel", "215"},
	{"215", "movie/title", "Aliens"},
	{"215", "movie/year", "1986"},
	{"215", "movie/director", "100"},
	{"215", "movie/cast", "139"},
	{"215", "movie/cast", "141"},
	{"215", "movie/cast", "103"},
	{"216", "movie/title", "Mad Max"},
	{"216", "movie/year", "1979"},
	{"216", "movie/director", "142"},
	{"216", "movie/cast", "112"},
	{"216", "movie/cast", "143"},
	{"216", "movie/cast", "144"},
	{"216", "movie/sequel", "217"},
	{"217", "movie/title", "Mad Max 2"},
	{"217", "movie/year", "1981"},
	{"217", "movie/director", "142"},
	{"217", "movie/cast", "112"},
	{"217", "movie/cast", "145"},
	{"217", "movie/cast", "146"},
	{"217", "movie/sequel", "218"},
	{"218", "movie/title", "Mad Max Beyond Thunderdome"},
	{"218", "movie/year", "1985"},
	{"218", "movie/director", "user"},
	{"218", "movie/director", "147"},
	{"218", "movie/cast", "112"},
	{"218", "movie/cast", "148"},
	{"219", "movie/title", "Braveheart"},
	{"219", "movie/year", "1995"},
	{"219", "movie/director", "112"},
	{"219", "movie/cast", "112"},
	{"219", "movie/cast", "149"},
}
