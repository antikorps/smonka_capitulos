# Smonka!
El uso de este script únicamente debe utilizarse para el derecho a la copia privada. Cualquier otro uso es inadecuado y es bajo completa responsabilidad del usuario.
## Enlaces de descarga
[GNU/LInux amd64](https://github.com/antikorps/smonka_capitulos/raw/main/bin/smonka_linux.zip "GNU/LInux amd64")
[Windows amd64](https://github.com/antikorps/smonka_capitulos/raw/main/bin/smonka_windows.zip "Windows amd64")
## Lista de capítulos
Los capítulos han sido nombrados de acuerdo a la web de PlutoTV y su número corresponde al orden en el que aparecen listados.
Visita https://pluto.tv/es/on-demand/series/smonka-es/ para más detalles
## Usuarios de Windows
El script no necesita ningún tipo de instalación, solo ejecutar. En el caso de que se necesite utilizar la interfaz de texto en la terminal se recomienda ejecutar el script con el archivo **ejecutarScript.bat**.  Ejecutarlo de esta forma proporcionará que al acabar no se cierre la ventana directamente y se pueda revisar la información mostrada. 
## Usuarios de GNU/Linux
Tal vez tengas que dar permisos de ejecución a los binarios (script y ffmpeg). Puedes hacerlo de la siguiente forma:
```bash
sudo chmod +x smonka_capitulos
sudo chmod +x recursos/ffmpeg
```
Para la obtención de los vídeos se utiliza el ejecutable de FFmpeg preparado por johnvansickle.com. Este ejecutable presenta el siguiente problema:
> A limitation of statically linking glibc is the loss of DNS resolution. Installing nscd through your package manager will fix this.

Por lo tanto, para utilizar este script es necesario tener instalado nscd (el tamaño es ridículo, creo que no llega al mega). Para instalarlo en sistemas operativos derivados de Debian:
```bash
sudo apt-get install nscd
```
## Funcionamiento
Existen 2 opciones para interactuar, a través de la línea de comandos (CLI) o una interfaz de texto basada en la terminal (Terminal user interface o TUI).
### CLI 
Utilizar los siguientes argumentos:
```bash
./smonka_capitulos -capitulos 1,2,3 -destino /home/user/Descargas -espera 3

```
-capitulos: enumeración del identificador numérico de los capítulos a descargar separados por comas, ej: 1,2,3,5
-destino: ruta completa de la carpeta donde se guardarán los vídeos
-espera: tiempo en segundos que se esperará entre las descargas de cada capitulo
### TUI
El proceso estará guiado y permitirá escoger los capítulos a descargar gracias a un listado, la carpeta de destino utilizando un selector de archivos y, finalmente, el tiempo de espera entre las descargas de cada capítulo.
![TUI](https://i.imgur.com/p2DVpip.png "TUI")

![TUI2](https://i.imgur.com/aVceRC7.png "TUI2")