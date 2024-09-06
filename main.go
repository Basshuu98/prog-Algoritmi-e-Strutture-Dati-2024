package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Piastrella struct {
	x         int
	y         int
	colore    string
	intensita int
}

type Piano struct {
	piastrelle map[punto]*Piastrella // insieme delle piastrelle
	regole     *[]Regola             // insieme delle regole
}

// Elemento di una regola
type termine struct {
	k     int
	alpha string
}

type Regola struct {
	termini     []termine
	nuovoColore string
	consumo     int
}

// Punto che identifica i vertici di una piastrella nel piano
type punto struct {
	x int
	y int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	p := Piano{piastrelle: make(map[punto]*Piastrella), regole: &[]Regola{}}
	for scanner.Scan() {
		s := scanner.Text()
		esegui(p, s)
	}
}

// Applica al Piano p l’operazione associata alla stringa s
func esegui(p Piano, s string) {
	comando := strings.Fields(s)
	var x int
	var y int
	if len(comando) > 2 {
		x, _ = strconv.Atoi(comando[1])
		y, _ = strconv.Atoi(comando[2])
	}

	switch string(comando[0]) {
	case "C":
		alpha := comando[3]
		i, _ := strconv.Atoi(comando[4])
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
		p.intensitaBlocco(x, y, false)
	case "B":
		p.intensitaBlocco(x, y, true)
	case "p":
		p.propaga(x, y)
	case "P":
		p.propagaBlocco(x, y)
	case "o":
		p.ordina()
	case "i":
		x2, _ := strconv.Atoi(comando[3])
		y2, _ := strconv.Atoi(comando[4])
		p.minIntensita(x, y, x2, y2)
	case "m":
		fmt.Println(p.perimetro(x, y))
	case "q":
		os.Exit(0)
	default:
		fmt.Println("Comando non supportato")
	}
}

// Colora Piastrella(x, y) di colore alpha e intensità i, qualunque sia lo stato di Piastrella(x, y) prima dell’operazione
func colora(p Piano, x int, y int, alpha string, i int) {
	piastrella := p.restituisciPiastrella(x, y)
	if piastrella != nil {
		piastrella.intensita = i
		piastrella.colore = alpha
	} else { // piastrella non è presente nel Piano, quindi creo una nuova Piastrella
		nuovaPiastrella := &Piastrella{x, y, alpha, i}
		p.piastrelle[punto{x, y}] = nuovaPiastrella
	}
}

// Spegne Piastrella(x, y), mentre non fa nulla se è gia spenta
func spegni(p Piano, x int, y int) {
	if piastrella := p.restituisciPiastrella(x, y); piastrella != nil && piastrella.intensita != 0 {
		piastrella.intensita = 0
	}
}

// Stampa e restituisce il colore e l’intensità di Piastrella(x, y). Se è spenta, non stampa nulla e restituisce la stringa vuota e l’intero 0
func stato(p Piano, x int, y int) (string, int) {
	if piastrella := p.restituisciPiastrella(x, y); piastrella != nil && piastrella.intensita != 0 {
		fmt.Printf("%s %d\n", piastrella.colore, piastrella.intensita)
		return piastrella.colore, piastrella.intensita
	}
	return "", 0
}

// Definisce la regola di propagazione k1α1 + k2α2 + · · · + knαn → β e la inserisce in fondo all’elenco.
func regola(p Piano, r string) {
	elementi := strings.Fields(r)
	beta := elementi[1]
	var termini []termine
	for i := 2; i < len(elementi); i += 2 {
		k, _ := strconv.Atoi(elementi[i])
		termini = append(termini, termine{k, elementi[i+1]})
	}
	// Creo la regola e la inserisco in fondo all'elenco di regole
	*p.regole = append(*p.regole, Regola{termini, beta, 0})
}

// Stampa l’elenco delle regole di propagazione, nell’ordine attuale.
func stampa(p Piano) {
	fmt.Println("(")
	for _, regola := range *p.regole {
		fmt.Printf("%s: ", regola.nuovoColore)
		for i := 0; i < len(regola.termini)-1; i++ {
			fmt.Printf("%d %s ", regola.termini[i].k, regola.termini[i].alpha)
		}
		fmt.Printf("%d %s\n", regola.termini[len(regola.termini)-1].k, regola.termini[len(regola.termini)-1].alpha)
	}
	fmt.Println(")")
}

// Calcola e stampa la somma delle intensità delle piastrelle contenute nel blocco omogeneo o non omogeneo di appartenenza alla piastrella(x, y).
// Se Piastrella(x, y) è spenta, restituisce 0.
func (p Piano) intensitaBlocco(x int, y int, omog bool) {
	piastrella := p.restituisciPiastrella(x, y)
	if piastrella == nil || piastrella.intensita == 0 {
		fmt.Println(0)
		return
	}
	visitati := make(map[punto]bool)
	somma := 0
	// Faccio il calcolo specifico se il blocco è omogeneo oppure no
	if omog {
		p.dfs(piastrella, visitati, piastrella.colore, omog, nil, &somma)
	} else {
		p.dfs(piastrella, visitati, "", omog, nil, &somma)
	}
	fmt.Println(somma)
}

// Visita in profondita per esplorare il blocco di appartenenza della piastrella
// Calcola la somma delle intensita sia per blocco che per bloccoOmog richiesta in intensitaBlocco, e le piastrelle del blocco richiesto in propagaBlocco
func (p Piano) dfs(piastrella *Piastrella, visitati map[punto]bool, colore string, omogeneo bool, blocco map[punto]*Piastrella, somma *int) {
	// Controllo se la piastrella è spenta oppure se non ha il colore adeguato
	if piastrella.intensita == 0 || (omogeneo && piastrella.colore != colore) {
		return
	}
	visitati[punto{piastrella.x, piastrella.y}] = true
	// Trovo le piastrelle del blocco solo se necessario (per propagaBlocco)
	if blocco != nil {
		blocco[punto{piastrella.x, piastrella.y}] = piastrella
	}
	// Calcolo la somma solo se necessario (per intensitaBlocco)
	if somma != nil {
		*somma += piastrella.intensita
	}
	// esploro le piastrelle adiacenti non ancora visitate
	for _, adj := range p.piastrelleCirconvicine(piastrella.x, piastrella.y) {
		if !visitati[punto{adj.x, adj.y}] {
			p.dfs(adj, visitati, colore, omogeneo, blocco, somma)
		}
	}
}

// Applica a Piastrella(x, y) la prima regola di propagazione applicabile dell’elenco, ricolorando la piastrella.
// Se nessuna regola è applicabile, non viene eseguita alcuna operazione
func (p Piano) propaga(x int, y int) {
	piastrella := p.restituisciPiastrella(x, y)
	if regola := p.restituisciRegola(x, y); regola != nil {
		// Controllo se la piastrella c'è già e applico la regola, altrimenti la accendo
		if piastrella != nil {
			piastrella.colore = regola.nuovoColore
		} else {
			colora(p, x, y, regola.nuovoColore, 1)
		}
	}
}

// Propaga il colore sul blocco di appartenenza di Piastrella(x, y)
func (p Piano) propagaBlocco(x int, y int) {
	piastrella := p.restituisciPiastrella(x, y)
	if piastrella == nil {
		return
	}
	visitati := make(map[punto]bool)
	blocco := make(map[punto]*Piastrella)
	// Prendo le piastrelle del blocco di appartenenza della piastrella (x,y)
	p.dfs(piastrella, visitati, "", false, blocco, nil)
	// Mappa dei cambiamenti colore
	aggiornamenti := make(map[punto]string)
	for _, piastrella := range blocco {
		r := p.restituisciRegola(piastrella.x, piastrella.y)
		if r != nil {
			aggiornamenti[punto{piastrella.x, piastrella.y}] = r.nuovoColore
		}
	}
	for pos, nuovoColore := range aggiornamenti {
		piastrella := p.restituisciPiastrella(pos.x, pos.y)
		piastrella.colore = nuovoColore
	}
}

// Restituisce la prima regola da applicare data una piastrella, nil altrimenti,ed aggiorno il consumo della regola (se) restituita
func (p Piano) restituisciRegola(x int, y int) *Regola {
	intorno := make(map[string]int)
	// Calcolo l'intorno
	for _, vicino := range p.piastrelleCirconvicine(x, y) {
		if _, ok := intorno[vicino.colore]; ok {
			intorno[vicino.colore]++
		} else {
			intorno[vicino.colore] = 1
		}
	}
	// Trovo la regola da applicare se esiste
	for i := range *p.regole {
		applicabile := true
		for _, parte := range (*p.regole)[i].termini {
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

// Ordina per consumo
func (p Piano) ordina() {
	sort.SliceStable(*p.regole, func(i, j int) bool {
		return (*p.regole)[i].consumo < (*p.regole)[j].consumo
	})
}

// Implemento gli elementi necessari per usare la funzione basandosi sull'algoritmo di Dijkstra
// Elemento per la coda di priorità
type Elemento struct {
	piastrella *Piastrella
	priorita   int
	indice     int
}

// PriorityQueue implementa una coda di priorità
type PriorityQueue []*Elemento

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priorita < pq[j].priorita
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].indice = i
	pq[j].indice = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	elem := x.(*Elemento)
	elem.indice = n
	*pq = append(*pq, elem)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	elem := old[n-1]
	old[n-1] = nil
	elem.indice = -1
	*pq = old[0 : n-1]
	return elem
}

func (pq *PriorityQueue) update(elem *Elemento, priorita int) {
	elem.priorita = priorita
	heap.Fix(pq, elem.indice)
}

func (p Piano) minIntensita(inizioX, inizioY, fineX, fineY int) int {

	// Inanzitutto controllo che le due piastrelle esistano e che siano accese
	inizio := p.restituisciPiastrella(inizioX, inizioY)
	fine := p.restituisciPiastrella(fineX, fineY)
	if inizio == nil || fine == nil || inizio.intensita == 0 || fine.intensita == 0 {
		return -1
	}

	// Inizializzo le distanze
	dist := make(map[punto]int)
	for pos := range p.piastrelle {
		dist[pos] = int(^uint(0) >> 1) // Imposto distanza a infinito (MaxInt)
	}

	// Coda di priorità
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	dist[punto{inizioX, inizioY}] = inizio.intensita
	heap.Push(&pq, &Elemento{piastrella: inizio, priorita: inizio.intensita})

	for pq.Len() > 0 {
		currentElemento := heap.Pop(&pq).(*Elemento)
		current := currentElemento.piastrella

		// Se ho raggiunto la destinazione, restituisco il costo minimo
		if current.x == fineX && current.y == fineY {
			fmt.Println(dist[punto{fineX, fineY}])
			return dist[punto{fineX, fineY}]
		}

		// Uso la funzione piastrelleCirconvicine per ottenere i vicini
		for pos, vicini := range p.piastrelleCirconvicine(current.x, current.y) {
			newDist := dist[punto{current.x, current.y}] + vicini.intensita
			if newDist < dist[pos] {
				dist[pos] = newDist
				heap.Push(&pq, &Elemento{piastrella: vicini, priorita: newDist})
			}
		}
	}

	// Se non esiste un percorso valido
	return -1
}

// Restituisce la piastrella(x,y) dato x e y
func (p Piano) restituisciPiastrella(x int, y int) *Piastrella {
	if _, ok := p.piastrelle[punto{x, y}]; ok {
		return p.piastrelle[punto{x, y}]
	}
	return nil
}

// Restituisce le piastrelle circonvicine alla piastrella(x,y)
func (p Piano) piastrelleCirconvicine(x, y int) (vicini map[punto]*Piastrella) { // restituisce le piastrelle circonvicine data una piastrella (x,y)
	vicini = make(map[punto]*Piastrella)
	direzioni := []punto{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}, {x - 1, y - 1}, {x + 1, y - 1}, {x - 1, y + 1}, {x + 1, y + 1}}
	for _, adj := range direzioni {
		if vicino, ok := p.piastrelle[punto{adj.x, adj.y}]; ok {
			vicini[punto{vicino.x, vicino.y}] = vicino
		}
	}
	return
}

func (p Piano) perimetro(x, y int) int {

	piastrellaIniziale := p.restituisciPiastrella(x, y)
	if piastrellaIniziale == nil || piastrellaIniziale.intensita == 0 {
		// Se la piastrella iniziale non esiste o è spenta, il perimetro è 0
		return 0
	}
	// Mappa per tenere traccia delle piastrelle visitate
	visitati := make(map[punto]bool)

	// Funzione DFS per calcolare il perimetro
	var dfsPerimetro func(piastrella *Piastrella) int
	dfsPerimetro = func(piastrella *Piastrella) int {
		visitati[punto{piastrella.x, piastrella.y}] = true
		perimetro := 0

		// Ottiengo le piastrelle circonvicine (incluse quelle diagonali)
		vicini := p.piastrelleCirconvicine(piastrella.x, piastrella.y)

		latiNonCondivisi := 4

		// Per considerare i lati condivisi con altre piastrelle devo considerare le sole direzioni orizzontali e verticali e non quelle diagonali
		direzioniOrizzontaliVerticali := []punto{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

		for _, vicino := range vicini {
			if vicino != nil && vicino.intensita > 0 {
				isOrizzontaleVerticale := false
				for _, dir := range direzioniOrizzontaliVerticali {
					// Calcolo la punto relativa della piastrella vicina rispetto alla piastrella corrente per poter confrontare la direzione (se diagonale o no)
					posRelativa := punto{vicino.x - piastrella.x, vicino.y - piastrella.y}

					if posRelativa == dir {
						isOrizzontaleVerticale = true
						break
					}
				}

				// Se è una piastrella orizzontale o verticale, condividono un lato
				if isOrizzontaleVerticale {
					latiNonCondivisi--
				}

				// Se non ho ancora visitato il vicino, lo esploro
				if !visitati[punto{vicino.x, vicino.y}] {
					perimetro += dfsPerimetro(vicino)
				}
			}
		}

		// Aggiungo i lati non condivisi al perimetro totale
		perimetro += latiNonCondivisi

		return perimetro
	}

	// uso la DFS
	return dfsPerimetro(piastrellaIniziale)
}
