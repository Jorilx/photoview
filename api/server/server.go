package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"

	"github.com/viktorstrate/photoview/api/database"
	"github.com/viktorstrate/photoview/api/graphql/auth"

	"github.com/99designs/gqlgen/handler"
	photoview_graphql "github.com/viktorstrate/photoview/api/graphql"
)

const defaultPort = "4001"

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = defaultPort
	}

	db := database.SetupDatabase()
	defer db.Close()

	// Migrate database
	if err := database.MigrateDatabase(db); err != nil {
		log.Fatalf("Could not migrate database: %s\n", err)
	}

	router := chi.NewRouter()
	router.Use(auth.Middleware(db))

	graphqlResolver := photoview_graphql.Resolver{Database: db}
	graphqlDirective := photoview_graphql.DirectiveRoot{}
	graphqlDirective.IsAdmin = photoview_graphql.IsAdmin(db)

	graphqlConfig := photoview_graphql.Config{
		Resolvers:  &graphqlResolver,
		Directives: graphqlDirective,
	}

	router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	router.Handle("/query", handler.GraphQL(photoview_graphql.NewExecutableSchema(graphqlConfig)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}