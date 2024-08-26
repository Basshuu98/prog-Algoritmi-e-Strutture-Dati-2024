package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Piastrella struct { // componente del piano
	x         int
	y         int
	colore    string
	intensita int
}

type piano struct {
	piastrelle map[posizione]*Piastrella // insieme delle piastrelle
	regole     *[]Regola                 // insieme delle regole
}

type termine struct { // parte di una regola
	k     int
	alpha string
}

type Regola struct { // regola
	parti       []termine
	nuovoColore string
	consumo     int
}

type posizione struct { // posizione nel piano
	x int
	y int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	p := piano{piastrelle: make(map[posizione]*Piastrella), regole: &[]Regola{}}
	for scanner.Scan() {
		s := scanner.Text()
		esegui(p, s)
	}
}

// Applica al sistema rappresentato da p l’operazione associata dalla stringa s
func esegui(p piano, s string) {
	comandi := strings.Fields(s)
	var x int
	var y int
	if len(comandi) > 2 {
		x, _ = strconv.Atoi(comandi[1])
		y, _ = strconv.Atoi(comandi[2])
	}

	switch string(comandi[0]) {
	case "C":
		alpha := comandi[3]
		i, _ := strconv.Atoi(comandi[4])
		colora(p, x, y, alpha, i)
	case "S":
		spegni(p, x, y)
	case "r":
		regola(p, s)
	case "?":
		stato(p, x, y)
	case "s":
		stampa(p)
	case "b":
		p.bloccoGenerale(x, y, false)
	case "B":
		p.bloccoGenerale(x, y, true)
	case "p":
		p.propaga(x, y)
	case "P":
		p.propagaBlocco(x, y)
	case "o":
		p.ordina()
	case "q":
		os.Exit(0)
	case "t":
		p.pista(x, y, s[6:])
	case "L":
		x2, _ := strconv.Atoi(comandi[3])
		y2, _ := strconv.Atoi(comandi[4])
		p.lung(x, y, x2, y2)
	default:
		fmt.Println("Comando non riconosciuto")
	}
}

// Colora Piastrella(x, y) di colore alpha e intensità i, qualunque sia lo stato di Piastrella(x, y) prima dell’operazione
func colora(p piano, x int, y int, alpha string, i int) {
	if piastrella := p.restituisciPiastrella(x, y); piastrella != nil { // piastrella gia' presente nel piano
		piastrella.intensita = i
		piastrella.colore = alpha
	} else { // piastrella non presente nel piano
		nuovaPiastrella := &Piastrella{x, y, alpha, i} // creo una piastrella nuova
		p.piastrelle[posizione{x, y}] = nuovaPiastrella
	}
}

// Spegne Piastrella(x, y). Se Piastrella(x, y) è già spenta, non fa nulla.
func spegni(p piano, x int, y int) {
	if piastrella := p.restituisciPiastrella(x, y); piastrella != nil && piastrella.intensita != 0 {
		piastrella.intensita = 0
	}
}

// Stampa e restituisce il colore e l’intensità di Piastrella(x, y). Se Piastrella(x, y) è spenta, non stampa nulla e restituisce la stringa vuota e l’intero 0
func stato(p piano, x int, y int) (string, int) {
	if piastrella := p.restituisciPiastrella(x, y); piastrella != nil && piastrella.intensita != 0 {
		fmt.Printf("%s %d\n", piastrella.colore, piastrella.intensita)
		return piastrella.colore, piastrella.intensita
	}
	return "", 0
}

// Definisce la regola di propagazione k1 α1 + k2 α2 + · · · + kn αn → β e la inserisce in fondo all’elenco delle regole.
func regola(p piano, r string) {
	parti := strings.Fields(r)
	beta := parti[1]
	var termini []termine
	for i := 2; i < len(parti); i += 2 {
		k, _ := strconv.Atoi(parti[i])
		termini = append(termini, termine{k, parti[i+1]})
	}
	*p.regole = append(*p.regole, Regola{termini, beta, 0}) // creo la regola e la inserisco in fondo all'elenco di regole
}

// Stampa l’elenco delle regole di propagazione, nell’ordine attuale.
func stampa(p piano) {
	fmt.Println("(")
	for _, regola := range *p.regole {
		fmt.Printf("%s: ", regola.nuovoColore)
		for i := 0; i < len(regola.parti)-1; i++ {
			fmt.Printf("%d %s ", regola.parti[i].k, regola.parti[i].alpha)
		}
		fmt.Printf("%d %s\n", regola.parti[len(regola.parti)-1].k, regola.parti[len(regola.parti)-1].alpha) // per ultima parte non metto lo spazio
	}
	fmt.Println(")")
}

// Calcola e stampa la somma delle intensità delle piastrelle contenute nel blocco omogeneo e non di appartenenza di piastrella(x, y).
// Se Piastrella(x, y) è spenta, restituisce 0.
func (p piano) bloccoGenerale(x int, y int, omog bool) {
	piastrella := p.restituisciPiastrella(x, y)
	if piastrella == nil || piastrella.intensita == 0 {
		fmt.Println(0)
		return
	}
	visite := make(map[posizione]bool)
	somma := 0
	if omog {
		p.dfs(piastrella, visite, piastrella.colore, omog, nil, &somma) // calcola bloccoOmog
	} else {
		p.dfs(piastrella, visite, "", omog, nil, &somma) // calcola blocco
	}
	fmt.Println(somma)
}

// Visita in profondita per esplorare il blocco di appartenenza della piastrella
// Calcola la somma delle intensita sia per blocco che per bloccoOmog, e le piastrelle del blocco usato in propagaBlocco
func (p piano) dfs(piastrella *Piastrella, visite map[posizione]bool, colore string, omogeneo bool, blocco map[posizione]*Piastrella, somma *int) {
	if piastrella.intensita == 0 || (omogeneo && piastrella.colore != colore) { // se piastrella spenta oppure non ha il colore adeguato
		return
	}
	visite[posizione{piastrella.x, piastrella.y}] = true

	if blocco != nil { // trovo le piastrelle del blocco solo se necessario (per propagaBlocco)
		blocco[posizione{piastrella.x, piastrella.y}] = piastrella
	}
	if somma != nil { // calcolo la somma solo se necessario (per bloccoGenerale)
		*somma += piastrella.intensita
	}

	for _, adj := range p.piastrelleCirconvicine(piastrella.x, piastrella.y) { // esploro i nodi adiacenti non ancora visitati
		if !visite[posizione{adj.x, adj.y}] {
			p.dfs(adj, visite, colore, omogeneo, blocco, somma)
		}
	}
}

// Applica a Piastrella(x, y) la prima regola di propagazione applicabile dell’elenco, ricolorando la piastrella.
// Se nessuna regola è applicabile, non viene eseguita alcuna operazione
func (p piano) propaga(x int, y int) {
	piastrella := p.restituisciPiastrella(x, y)
	if regola := p.restituisciRegola(x, y); regola != nil {
		if piastrella != nil { // la piastrella esisteva gia
			piastrella.colore = regola.nuovoColore
		} else { // la piastrella non esisteva == spenta
			colora(p, x, y, regola.nuovoColore, 1)
		}
	}
}

// Propaga il colore sul blocco di appartenenza di Piastrella(x, y)
func (p piano) propagaBlocco(x int, y int) {
	piastrella := p.restituisciPiastrella(x, y)
	if piastrella == nil { // se piastrella spenta non faccio nulla
		return
	}
	visite := make(map[posizione]bool)
	blocco := make(map[posizione]*Piastrella)
	p.dfs(piastrella, visite, "", false, blocco, nil) // prendo le piastrelle del blocco di appartenenza della piastrella (x,y)
	aggiornamenti := make(map[posizione]string)       // mappa dei cambiamenti colore
	for _, piastrella := range blocco {
		r := p.restituisciRegola(piastrella.x, piastrella.y)
		if r != nil {
			aggiornamenti[posizione{piastrella.x, piastrella.y}] = r.nuovoColore
		}
	}
	for pos, nuovoColore := range aggiornamenti { // applico le regole
		piastrella := p.restituisciPiastrella(pos.x, pos.y)
		piastrella.colore = nuovoColore
	}
}

// Restituisce la prima regola da applicare data una piastrella, nil altrimenti
// Aggiorna il consumo della regola restituita
func (p piano) restituisciRegola(x int, y int) *Regola {
	intorno := make(map[string]int)
	for _, vicino := range p.piastrelleCirconvicine(x, y) { // calcolo l'intorno
		if _, ok := intorno[vicino.colore]; ok {
			intorno[vicino.colore]++
		} else {
			intorno[vicino.colore] = 1
		}
	}
	for i := range *p.regole { // trovo la regola da applicare se esiste
		applicabile := true
		for _, parte := range (*p.regole)[i].parti { // itero sulle parti della regola
			if occorrenzeIntorno, ok := intorno[parte.alpha]; ok {
				if occorrenzeIntorno < parte.k {
					applicabile = false
					break
				}
			} else {
				applicabile = false
				break
			}
		}
		if applicabile {
			(*p.regole)[i].consumo++
			return &(*p.regole)[i]
		}
	}
	return nil
}

// Ordina per consumo (stabile)
func (p piano) ordina() {
	sort.SliceStable(*p.regole, func(i, j int) bool {
		return (*p.regole)[i].consumo < (*p.regole)[j].consumo
	})
}

// Stampa la pista che parte da Piastrella(x, y) e segue la sequenza di direzioni s, se tale pista è definita.
// Altrimenti non stampa nulla
func (p piano) pista(x int, y int, s string) {
	piastrella := p.restituisciPiastrella(x, y)
	if piastrella == nil { // piastrella non presente (spenta)
		return
	}

	res := fmt.Sprintf("[\n%d %d %s %d\n", piastrella.x, piastrella.y, piastrella.colore, piastrella.intensita)
	comandi := strings.Split(s, ",")
	direzioni := map[string]posizione{"NN": posizione{0, 1}, "NE": posizione{1, 1}, "NO": posizione{-1, 1}, "EE": posizione{1, 0}, "SE": posizione{1, -1}, "SO": posizione{-1, -1}, "SS": posizione{0, -1}, "OO": posizione{-1, 0}}
	curr := piastrella // piastrella "corrente"
	stampa := true
	for _, comando := range comandi {
		movimento := direzioni[comando]
		next := p.restituisciPiastrella(curr.x+movimento.x, curr.y+movimento.y) // cerco prossima piastrella nella pista secondo il comando
		if next == nil || next.intensita == 0 {
			res += fmt.Sprintf("]")
			stampa = false
			return // pista non esiste in quanto la piastrella non e' presente oppure e' spenta
		}
		res += fmt.Sprintf("%d %d %s %d\n", next.x, next.y, next.colore, next.intensita)
		curr = next
	}
	res += fmt.Sprintf("]")
	if stampa {
		fmt.Println(res)
	}
}

// Determina la lunghezza della pista più breve che parte da Piastrella(x1 , y1 ) e arriva in Piastrella(x2 , y2 ).
// Altrimenti non stampa nulla.
func (p piano) lung(x1 int, y1 int, x2 int, y2 int) int {
	from := p.restituisciPiastrella(x1, y1) // controllo che le due piastrelle esistano e siano accese
	to := p.restituisciPiastrella(x2, y2)
	if from == nil || to == nil || from.intensita == 0 || to.intensita == 0 { // se spente o "non presenti nel piano"
		return -1
	}

	visite := make(map[posizione]bool)  // tengo traccia dei vertici gia' visitati
	distance := make(map[posizione]int) // tengo traccia della distanza dal vertice di partenza
	coda := []*Piastrella{from}         // creo la coda inserendo il vertice di partenza
	visite[posizione{from.x, from.y}] = true
	distance[posizione{from.x, from.y}] = 1

	for len(coda) > 0 { // applico BFS
		u := coda[0]    // dequeue() estraggo il primo elemento dalla coda
		coda = coda[1:] // dequeue() aggiorno la coda
		for _, vicino := range p.piastrelleCirconvicine(u.x, u.y) {
			if !visite[posizione{vicino.x, vicino.y}] { // se non ho ancora visitato il vicino allora lo aggiungo alla coda
				coda = append(coda, vicino) // enqueue() inserisco in fondo
				visite[posizione{vicino.x, vicino.y}] = true
				distance[posizione{vicino.x, vicino.y}] = distance[posizione{u.x, u.y}] + 1 // aggiorno distanza
			}
			if vicino.x == x2 && vicino.y == y2 {
				fmt.Println(distance[posizione{vicino.x, vicino.y}]) // lunghezza minima
				return distance[posizione{vicino.x, vicino.y}]
			}
		}
	}
	return -1
}

// Restituisce la piastrella(x,y) dato x e y
func (p piano) restituisciPiastrella(x int, y int) *Piastrella {
	if _, ok := p.piastrelle[posizione{x, y}]; ok {
		return p.piastrelle[posizione{x, y}]
	}
	return nil
}

// Restituisce le piastrelle circonvicine alla piastrella(x,y)
func (p piano) piastrelleCirconvicine(x, y int) (vicini map[posizione]*Piastrella) { // restituisce le piastrelle circonvicine data una piastrella (x,y)
	vicini = make(map[posizione]*Piastrella)
	direzioni := []posizione{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}, {x - 1, y - 1}, {x + 1, y - 1}, {x - 1, y + 1}, {x + 1, y + 1}}
	for _, adj := range direzioni {
		if vicino, ok := p.piastrelle[posizione{adj.x, adj.y}]; ok {
			vicini[posizione{vicino.x, vicino.y}] = vicino
		}
	}
	return
}
