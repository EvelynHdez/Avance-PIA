package main //Se intentó usar Daemon para que se compilara mientras se guardaba y se actualizaba

import (
	//importamos para el json
	"encoding/json"
	"fmt"

	//manera de manejar las entradas y las salidas de los datos que van llegando al servidor
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	//nos permite crear un servidor para la API
	"github.com/gorilla/mux"

	//Base de datos conexion
	"database/sql" // Interactuar con bases de datos

	_ "github.com/go-sql-driver/mysql" // La librería que nos permite conectar a MySQL
)

//Definimos las variables del tipo de dato "task"
//Se le responderá a la busqueda con estos datos
type task struct {
	ID     int    `json:ID`
	Nombre string `json:Nombre`
	Grupo  string `json:Grupo`
}

type Artista struct {
	IdA     int    `json:IdA`
	NombreA string `json:NombreA`
	GrupoA  string `json:GrupoA`
}

//Arreglo de cada una de las acciones "task"
type allTasks []task

//variable a definir la lista de tareas
var tasks = allTasks{ //variables que vamos a modificar
	{
		ID:     1,
		Nombre: "Task One",
		Grupo:  "Some Grupo",
	},
}

func obtenerBaseDeDatos() (db *sql.DB, e error) {
	usuario := "root"
	pass := ""
	host := "tcp(127.0.0.1:3306)"
	nombreBaseDeDatos := "pia2"
	// Debe tener la forma usuario:contraseña@host/nombreBaseDeDatos
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", usuario, pass, host, nombreBaseDeDatos))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func insertar(c Artista) (e error) {
	db, err := obtenerBaseDeDatos()
	if err != nil {
		return err
	}
	defer db.Close()

	c.IdA = len(tasks) + 1

	// Preparamos para prevenir inyecciones SQL
	sentenciaPreparada, err := db.Prepare("INSERT INTO artista (id, nombre, grupo) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer sentenciaPreparada.Close()
	// Ejecutar sentencia, un valor por cada '?'
	_, err = sentenciaPreparada.Exec(c.IdA, c.NombreA, c.GrupoA)
	if err != nil {
		return err
	}
	return nil

}

//ruta para obtener todas las tareas "task"
func getTasks(w http.ResponseWriter, r *http.Request) { //nos regresará las variables
	//Para cuando envie informacion sepa que lo estoy enviando en formato json
	//insertar(),
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

//funcion para crear tareas
func createTask(w http.ResponseWriter, r *http.Request) {
	//instancia de nuestro tipo de dato "task"
	var newTask task
	//La informacion que el usuario nos estará mandando al servidor
	reqBody, err := ioutil.ReadAll(r.Body)
	//validamos si hay un error
	if err != nil {
		fmt.Fprintf(w, "Insert a Valid Task")
	}
	//manipulamos la infromacion enviada
	json.Unmarshal(reqBody, &newTask)
	//ID autogenereado para cada nueva tarea
	newTask.ID = len(tasks) + 1
	//Devuelve la nueva tarea creada
	tasks = append(tasks, newTask)

	//tipo de informacion que estaremos mandando
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	//Se responde al usuario con la nueva tarea
	json.NewEncoder(w).Encode(newTask)

}

//Busqueda de la tarea
func getTask(w http.ResponseWriter, r *http.Request) {
	//extrae las variables desde el request
	vars := mux.Vars(r)
	//recibe un string y lo convierte en entero
	taskID, err := strconv.Atoi(vars["id"])

	//mensaje en caso de que haya algun error
	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	//recorre elemento a elemento de la lista de tareas
	for _, task := range tasks {
		if task.ID == taskID {
			w.Header().Set("Content-Type", "application/json")
			//Se enviaran las tareas con el json
			json.NewEncoder(w).Encode(task)
		}
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	for i, t := range tasks {
		if t.ID == taskID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			fmt.Fprintf(w, "The task with ID %v has been remove succesfully", taskID)

		}
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	var updateTask task

	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please enter valid data")
	}
	json.Unmarshal(reqBody, &updateTask)

	for i, t := range tasks {
		if t.ID == taskID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			updateTask.ID = taskID
			tasks = append(tasks, updateTask)

			fmt.Fprintf(w, "The task with ID %v has been updated successfully", taskID)
		}
	}

}

//w, las respuestas que le daremos al cliente
//r, la info que el usurio a mandado
func indexRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bienvenido a mi API")
}

//CORS
func enableCORS(router *mux.Router) {
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
	}).Methods(http.MethodOptions)
	router.Use(middlewareCors)
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			// Just put some headers to allow CORS...
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			// and call next handler!
			next.ServeHTTP(w, req)
		})
}

func main() {
	db, err := obtenerBaseDeDatos()
	if err != nil {
		fmt.Printf("Error obteniendo base de datos: %v", err)
		return
	}
	// Terminar conexión al terminar función
	defer db.Close()

	// Ahora vemos si tenemos conexión
	err = db.Ping()
	if err != nil {
		fmt.Printf("Error conectando: %v", err)
		return
	}
	// Listo, aquí ya podemos usar a db!
	fmt.Printf("Conectado correctamente")

	//prueba insert
	c := Artista{
		IdA:     9,
		NombreA: "Luis Cabrera Benito",
		GrupoA:  "Calle Sin Nombre #12",
	}
	err = insertar(c)
	if err != nil {
		fmt.Printf("Error insertando: %v", err)
	} else {
		fmt.Println("Insertado correctamente")
	}
	//

	//creamos un router con Mux y la guardamos con la variable "router"
	router := mux.NewRouter().StrictSlash(true)
	//ruta para cuando le demos clic a la url nos lleve a "index route"
	router.HandleFunc("/", indexRoute)
	//ruta para cuando pida la ruta "tasks" respondamos con"getTasks"
	router.HandleFunc("/tasks", getTasks).Methods("GET")
	//metodo para mandar la info a postman con el metodo post
	router.HandleFunc("/tasks", createTask).Methods("POST")
	//metodo para decirle que la Id va a cambiar
	router.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")

	//creamos el puerto para el http
	log.Fatal(http.ListenAndServe(":3000", router))

	enableCORS(router)
}
