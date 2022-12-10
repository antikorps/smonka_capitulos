package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ncruces/zenity"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

type ManejadorDescargas struct {
	BaseDatos            *sql.DB
	DescargasCorrectas   map[int]string
	DescargasIncorrectas map[int]string
	Capitulos            []int
	Errores              []string
	EsperaDescargas      int
	RutaDestino          string
	RutaRecursos         string
	RutaFFmpeg           string
}

func nuevoManejador() ManejadorDescargas {
	descargasCorrectas := make(map[int]string)
	descargasIncorrectas := make(map[int]string)
	return ManejadorDescargas{
		DescargasCorrectas:   descargasCorrectas,
		DescargasIncorrectas: descargasIncorrectas,
	}
}

func (m *ManejadorDescargas) configuracionBasica() {
	ejecutable, ejecutableError := os.Executable()
	if ejecutableError != nil {
		log.Fatalln(ejecutableError)
	}
	rutaEjecutable := filepath.Dir(ejecutable)
	rutaRecursos := filepath.Join(rutaEjecutable, "recursos")
	m.RutaRecursos = rutaRecursos

	if runtime.GOOS == "windows" {
		m.RutaFFmpeg = filepath.Join(rutaRecursos, "ffmpeg.exe")
	} else {
		m.RutaFFmpeg = filepath.Join(rutaRecursos, "ffmpeg")
	}

	rutaBaseDatos := filepath.Join(rutaRecursos, "bbdd.sqlite")
	conexionBaseDatos, conexionBaseDatosError := sql.Open("sqlite", rutaBaseDatos)
	if conexionBaseDatosError != nil {
		log.Fatalln(conexionBaseDatosError)
	}
	m.BaseDatos = conexionBaseDatos
}

func obtenerCapitulos(capitulos *string) []int {
	var capitulosValidos []int
	if *capitulos == "" {
		return capitulosValidos
	}
	caps := strings.Split(*capitulos, ",")
	for _, c := range caps {
		c := strings.TrimSpace(c)
		numero, numeroError := strconv.Atoi(c)
		if numeroError != nil {
			log.Println("número de capítulo no válido", numero, numeroError)
			continue
		}
		capitulosValidos = append(capitulosValidos, numero)
	}
	return capitulosValidos
}

func comprobarRutaDestino(destino *string) string {
	var rutaDestino string

	_, destinoError := os.Stat(*destino)
	if destinoError == nil {
		rutaDestino = *destino
	}
	return rutaDestino

}

func (m *ManejadorDescargas) entradaArgumentos(capitulos, destino *string, espera *int) {
	m.Capitulos = obtenerCapitulos(capitulos)
	m.RutaDestino = comprobarRutaDestino(destino)
	if *espera != 1 {
		m.EsperaDescargas = *espera
	}
}

func pintarMensajeBienvenida() {
	fmt.Println("")
	titulo, _ := pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString("SMONKA")).Srender()
	pterm.DefaultCenter.Println(titulo)
	fmt.Println("\nLos capítulos tienen el nombre de acuerdo a la web de PlutoTV y su nombre corresponde al orden en el que aparecen listados.\nVisita https://pluto.tv/es/on-demand/series/smonka-es/ para más detalles")
}

func solicitarCapitulosTui(baseDatos *sql.DB) []int {
	consulta := "SELECT id, nombre FROM smonka ORDER BY id ASC"
	filas, filasError := baseDatos.Query(consulta)
	if filasError != nil {
		log.Fatalln(filasError)
	}
	var relacionCapitulos []string
	var numeroResultados int
	for filas.Next() {
		numeroResultados++
		var identificador int
		var nombre string
		filas.Scan(&identificador, &nombre)

		opcion := fmt.Sprint("Capítulo: ", identificador, " ### Nombre capítulo:", nombre)
		relacionCapitulos = append(relacionCapitulos, opcion)
	}

	selector := pterm.DefaultInteractiveMultiselect.WithOptions(relacionCapitulos)
	selector.Filter = false
	mensajeDescarga := fmt.Sprint("Hay un total de ", numeroResultados, " capítulos\nMuevete con el cursor\nSelecciona con intro\nConfirma con tabulador")
	selector.DefaultText = mensajeDescarga
	selector.MaxHeight = 15
	opcionesSeleccionadas, opcionesSeleccionadasError := selector.Show()
	if opcionesSeleccionadasError != nil {
		log.Fatalln(opcionesSeleccionadasError)
	}

	var capitulosSeleccionados []int
	for _, opc := range opcionesSeleccionadas {
		expRegNumeroCapitulo := regexp.MustCompile(`Capítulo:\s(\d{1,3}).*`)
		opc = expRegNumeroCapitulo.ReplaceAllString(opc, "$1")
		numero, numeroError := strconv.Atoi(opc)
		if numeroError != nil {
			log.Println(numeroError)
			continue
		}
		capitulosSeleccionados = append(capitulosSeleccionados, numero)
	}
	pterm.Info.Printfln("Capítulos seleccionados: %s", pterm.Green(capitulosSeleccionados))
	return capitulosSeleccionados

}

func solicitarDestinoTui() string {
	var rutaDestino string
	pterm.Info.Printfln("Utiliza el selector de archivos que se abrirá a continuación para escoger el directorio guardar los vídeos")

	time.Sleep(2 * time.Second)

	rutaDestino, rutaDestinoError := zenity.SelectFile(
		zenity.Directory())
	if rutaDestinoError != nil {
		log.Fatalln(rutaDestinoError)
	}
	pterm.Info.Printfln("Directorio de descarga seleccionado: %s", pterm.Green(rutaDestino))
	return rutaDestino

}

func solicitarTiempoEsperaTui() int {
	var tiempoEspera int

	entradaUsuario := pterm.DefaultInteractiveTextInput
	entradaUsuario.DefaultText = "Introduce el número de segundos de espera entre cada descarga (en blanco es 0)"
	esperaUsuario, esperaUsuarioError := entradaUsuario.Show()
	if esperaUsuarioError != nil {
		log.Fatalln(esperaUsuarioError)
	}
	esperaUsuario = strings.TrimSpace(esperaUsuario)
	numero, numeroError := strconv.Atoi(esperaUsuario)
	if numeroError == nil {
		tiempoEspera = numero
	}

	return tiempoEspera
}

func (m *ManejadorDescargas) entradaTUI() {
	if len(m.Capitulos) > 0 && m.RutaDestino != "" && m.EsperaDescargas != -1 {
		return
	}
	pintarMensajeBienvenida()
	if len(m.Capitulos) == 0 {
		m.Capitulos = solicitarCapitulosTui(m.BaseDatos)
		if len(m.Capitulos) == 0 {
			os.Exit(0)
		}
	}
	if m.RutaDestino == "" {
		m.RutaDestino = solicitarDestinoTui()
		if m.RutaDestino == "" {
			os.Exit(0)
		}
	}
	if m.EsperaDescargas == -1 {
		m.EsperaDescargas = solicitarTiempoEsperaTui()
	}
}

func (m *ManejadorDescargas) descargarCapitulos() {
	for _, capitulo := range m.Capitulos {
		consulta := "SELECT m3u8 FROM smonka WHERE id = ?"
		filas, filasError := m.BaseDatos.Query(consulta, capitulo)
		if filasError != nil {
			m.DescargasIncorrectas[capitulo] = filasError.Error()
			pterm.Error.Println(filasError.Error())
			continue
		}
		for filas.Next() {
			var m3u8 string
			filas.Scan(&m3u8)

			nombreArchivo := fmt.Sprint(capitulo, "_smonka.m3u8")
			rutaArchivo := filepath.Join(m.RutaRecursos, nombreArchivo)
			nombreVideo := fmt.Sprint(capitulo, "_smonka.mp4")
			rutaVideo := filepath.Join(m.RutaDestino, nombreVideo)

			m3u8Archivo, m3u8ArchivoError := os.Create(rutaArchivo)
			if m3u8ArchivoError != nil {
				m.DescargasIncorrectas[capitulo] = m3u8ArchivoError.Error()
				pterm.Error.Println(m3u8ArchivoError.Error())
				continue
			}
			defer os.Remove(rutaArchivo)

			_, escrituraArchivoError := fmt.Fprint(m3u8Archivo, m3u8)
			if escrituraArchivoError != nil {
				m.DescargasIncorrectas[capitulo] = escrituraArchivoError.Error()
				pterm.Error.Println(escrituraArchivoError.Error())
				continue
			}

			mensajeDescarga := fmt.Sprint("Comienza el proceso de descarga del capítulo ", capitulo, ". Por favor, espera...")
			pterm.Info.Printfln(mensajeDescarga)

			comando := exec.Command(m.RutaFFmpeg, "-analyzeduration", "100M", "-probesize", "100M", "-protocol_whitelist", "file,http,https,tcp,tls,crypto", "-i", rutaArchivo, "-c", "copy", rutaVideo)

			ejecucion, ejecucionError := comando.CombinedOutput()
			if ejecucionError != nil {
				pterm.Error.Println(string(ejecucion))
				continue
			}

			mensajeExito := fmt.Sprint("Capítulo descargado con éxito en ", rutaVideo)
			pterm.Success.Println(mensajeExito)

			m.DescargasCorrectas[capitulo] = rutaVideo
			time.Sleep(time.Duration(m.EsperaDescargas))
		}
	}
}

func (m *ManejadorDescargas) resumenDescargas() {
	fmt.Println("")
	titulo, _ := pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString("RESUMEN")).Srender()
	pterm.DefaultCenter.Println(titulo)

	for c, v := range m.DescargasIncorrectas {
		mensajeError := fmt.Sprint("capítulo ", c, "no descargado: ", v)
		pterm.Error.Println(mensajeError)
	}

	for c, v := range m.DescargasCorrectas {
		mensajeError := fmt.Sprint("capítulo ", c, " descargado correctamente en: ", v)
		pterm.Success.Println(mensajeError)
	}
}

func main() {
	manejador := nuevoManejador()
	manejador.configuracionBasica()

	capitulos := flag.String("capitulos", "", "enumeración del identificador numérico de los capítulos a descargar separados por comas, ej: 1,2,3,5")
	destino := flag.String("destino", "", "ruta completa de la carpeta donde se guardarán los vídeos resultados")
	espera := flag.Int("espera", -1, "tiempo en segundos que se esperará entre las descargas de cada capitulo")
	flag.Parse()
	manejador.entradaArgumentos(capitulos, destino, espera)

	manejador.entradaTUI()

	manejador.descargarCapitulos()

	manejador.resumenDescargas()
}
