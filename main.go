package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/buaazp/fasthttprouter"
	_ "github.com/lib/pq"
	"github.com/likexian/whois-go"
	"github.com/likexian/whois-parser-go"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

// DomainRequest son los parametros del endpoint /domino
type DomainRequest struct {
	Domain string `json:"domain"`
}

// ServerResponse Parte de la respuesta al endpoint /dominio
type ServerResponse struct {
	Address  string `json:"address"`
	SslGrade string `json:"ssl_grade"`
	Country  string `json:"country"`
	Owner    string `json:"owner"`
}

// DomainResponse La respuesta del endpoint /dominio
type DomainResponse struct {
	Servers          []ServerResponse `json:"servers"`
	ServersChanged   bool             `json:"servers_changed"`
	SslGrade         string           `json:"ssl_grade"`
	PreviousSslGrade string           `json:"previous_ssl_grade"`
	Logo             string           `json:"logo"`
	Title            string           `json:"title"`
	IsDown           bool             `json:"is_down"`
}

// SSLLabsResponseEndpoint Parte de la respuesta de ssllabs
type SSLLabsResponseEndpoint struct {
	IPAddress         string `json:"ipAddress"`
	StatusMessage     string `json:"statusMessage"`
	Grade             string `json:"grade"`
	GradeTrustIgnored string `json:"gradeTrustIgnored"`
	HasWarnings       bool   `json:"hasWarnings"`
	IsExceptional     bool   `json:"isExceptional"`
	Progress          int    `json:"progress"`
	Duration          int    `json:"duration"`
	Delegation        int    `json:"delegation"`
}

// SSLLabsResponse Respuesta ala petición de la api de SSLLabs
type SSLLabsResponse struct {
	Host            string                    `json:"host"`
	Port            int                       `json:"port"`
	Protocol        string                    `json:"protocol"`
	IsPublic        bool                      `json:"isPublic"`
	Status          string                    `json:"status"`
	StatusMessage   string                    `json:"statusMessage"`
	StartTime       int                       `json:"startTime"`
	TestTime        int                       `json:"testTime"`
	EngineVersion   string                    `json:"engineVersion"`
	CriteriaVersion string                    `json:"criteriaVersion"`
	Endpoints       []SSLLabsResponseEndpoint `json:"endpoints"`
}

// DatabaseServesRow La estructura de la table servers
type DatabaseServesRow struct {
	ID        int
	Address   string
	SslGrade  string
	Country   string
	Owner     string
	CreatedAt time.Time
	DomainID  int
}

// DatabaseDomainRow La estrctura de la tabla domains
type DatabaseDomainRow struct {
	ID    int
	Title string
	Logo  string
	Host  string
}

var db *sql.DB // Database connection pool.

func index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Endpoints: \n /dominios")
}

func getSslAndServerInformation(domain string) (SSLLabsResponse, error) {
	fmt.Printf("getSslAndServerInformation: Dominio solicitado es %s\n", domain)
	var parsedResponse SSLLabsResponse
	var err error
	// Armamos la url
	endpoint, err := url.ParseRequestURI("https://api.ssllabs.com/api/v3/analyze?host=" + domain)
	// Si es fallo tiramos
	if err != nil {
		return parsedResponse, err
	}

	log.Println(endpoint.String())
	// Solicitamos la url
	res, err := http.Get(endpoint.String())

	// Si es fallo tiramos
	if err != nil {
		return parsedResponse, err
	}

	// Leemos el json
	decoder := json.NewDecoder(res.Body)
	decoderErr := decoder.Decode(&parsedResponse)
	if decoderErr != nil {
		log.Fatalf("Error procesando la respuesta de ssllabs %v", err)
	}
	// Cerramos la petición cuando haya sido leida la respuesta
	defer res.Body.Close()

	return parsedResponse, err
}

func boostrapAndGetDatabase() (*sql.DB, error) {
	var err error
	db, err := sql.Open("postgres", "postgresql://test@localhost:26257/domains?sslmode=disable")
	if err != nil {
		return db, err
	}

	// Crea la tabla de servidores
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS servers (
        id SERIAL PRIMARY KEY,
        address STRING,
        ssl_grade STRING,
        country STRING,
        owner STRING,
        domain_id INT,
        created_at TIMESTAMP
      )`)
	if err != nil {
		return db, err
	}
	// Crea la tabla de dominios
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS domains (
        id SERIAL PRIMARY KEY,
        title STRING,
        logo STRING,
        host STRING
      )`)
	if err != nil {
		return db, err
	}

	if err := db.Ping(); err != nil {
		log.Panic(err)
	}
	return db, err
}

func insertIntoDomain(payload *DatabaseDomainRow) (DatabaseDomainRow, error) {
	var ID int
	var err error
	var result DatabaseDomainRow
	sqlStatement := `INSERT INTO domains (title, logo, host)
    VALUES ($1, $2, $3)
    RETURNING ID`
	fmt.Printf("insertIntoDomain: Carga es %+v\n, SQL: %s\n", payload, sqlStatement)
	statement, err := db.Prepare(sqlStatement)
	if err != nil {
		return result, err
	}
	err = statement.QueryRow(
		payload.Title,
		payload.Logo,
		payload.Host).Scan(&ID)
	if err != nil {
		return result, err
	}
	sqlSelectNewRecordStatement := `SELECT idm title, logo, host FROM domains WHERE ID = $1`
	err = db.QueryRow(sqlSelectNewRecordStatement, ID).Scan(
		&result.ID,
		&result.Logo,
		&result.Title,
		&result.Host)
	if err != nil {
		return result, err
	}
	return result, err
}

func insertIntoServer(payload *DatabaseServesRow) (DatabaseServesRow, error) {
	var ID int
	var result DatabaseServesRow
	var err error
	sqlStatement := `INSERT INTO servers (address, ssl_grade, country, created_at, domain_id)
    VALUES ($1, $2, $3, $4, $5, $7)
    RETURNING ID`
	fmt.Printf("insertIntoServer: Carga es %+v\n, SQL: %s\n", payload, sqlStatement)
	err = db.QueryRow(
		sqlStatement,
		payload.Address,
		payload.SslGrade,
		payload.Country,
		payload.Owner,
		time.Now,
		payload.DomainID).Scan(&ID)
	if err != nil {
		return result, err
	}
	sqlSelectNewRecordStatement := `SELECT id, address, ssl_grade, country, owner, domain_id, created_at
    FROM servers WHERE ID = $1`
	err = db.QueryRow(sqlSelectNewRecordStatement, ID).Scan(
		&result.ID,
		&result.Address,
		&result.SslGrade,
		&result.Country,
		&result.DomainID,
		&result.CreatedAt)
	if err != nil {
		return result, err
	}
	return result, err
}

func getRowsFromServersByDomainID(domainID int) ([]DatabaseServesRow, error) {
	var result []DatabaseServesRow
	var err error
	sqlStatement := `
  SELECT id, address, ssl_grade, country, owner, domain_id, created_at
  FROM servers 
  WHERE domain_id = $1`
	fmt.Printf("getRowsFromServersByDomainID: Dominio solicitado es %d\n, SQL: %s\n", domainID, sqlStatement)

	rows, err := db.Query(sqlStatement, domainID)
	defer rows.Close()
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return result, err
		}
		return result, err
	}
	for rows.Next() {
		var row DatabaseServesRow

		if err := rows.Scan(&row.ID, &row.Address, &row.SslGrade, &row.Country, &row.Owner, &row.DomainID, &row.CreatedAt); err != nil {
			return result, err
		}
		result = append(result, row)
	}
	return result, err

}

func getRowFromDomainsByHost(domain string) (DatabaseDomainRow, error) {
	var result DatabaseDomainRow
	var err error
	sqlStatement := `SELECT id, host, title, logo FROM domains
    WHERE host = $1
    LIMIT 1`
	fmt.Printf("getRowFromDomainsByHost: Dominio solicitado es %s\n, SQL: %s\n", domain, sqlStatement)
	if err := db.Ping(); err != nil {
		return result, err
	}
	preparedStatement, prepareError := db.Prepare(sqlStatement)

	if prepareError != nil {
		return result, err
	}
	err = preparedStatement.QueryRow(domain).Scan(&result.ID, &result.Host, &result.Title, &result.Logo)
	if err != nil {
		return result, err
	}

	return result, err
}
func getRowFromDomainsByID(ID int) (DatabaseDomainRow, error) {
	sqlStatement := `SELECT id, title, logo FROM domains
    WHERE id = $1
    LIMIT 1`
	fmt.Printf("getRowFromDomainsByID: id %d\n, SQL: %s\n", ID, sqlStatement)
	var result DatabaseDomainRow
	err := db.QueryRow(sqlStatement, ID).Scan(&result.ID, &result.Logo, &result.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, err
		}
		log.Fatal(err)
	}
	return result, nil
}

func getLogoAndTitleFromDomain(domain string) (string, string, error) {
	var logo string
	var title string
	var err error
	// Request the HTML page.
	res, err := http.Get("https://" + domain)
	if err != nil {
		return logo, title, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("getLogoAndTitleFromDomain: status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return logo, title, err
	}

	// Find the logo
	linkEl := doc.Find(`link[rel*="icon"][href$="png"]`).First()
	logo, _ = linkEl.Attr("href")
	// Find the title
	titleEl := doc.Find("title").First()
	title = titleEl.Text()
	return logo, title, err

}

type whoisResult struct {
	country string
	owner   string
}

func getWhoisInfo(ipOrDomain string) (whoisResult, error) {
	var result whoisResult
	var err error
	whoisString, err := whois.Whois(ipOrDomain)
	if err != nil {
		return result, err
	}
	parsed, err := whoisparser.Parse(whoisString)
	if err != nil {
		fmt.Printf("[getWhoisInfo] error: %s; rawWhois: \n%s\n", err.Error(), whoisString)
		return result, err
	}
	fmt.Printf("parsedWhois %+v\n", parsed)
	result.country = parsed.Administrative.Organization
	result.owner = parsed.Administrative.Country
	return result, err
}

func getOrCreateServerDomainRecords(domain string) (DatabaseDomainRow, []DatabaseServesRow, bool, error) {
	fmt.Printf("getOrCreateServerDomainRecords: Dominio solicitado es %s\n", domain)
	var serverRecords []DatabaseServesRow
	var domainRecord DatabaseDomainRow
	serversHasChanged := false
	var err error
	sslInfo, err := getSslAndServerInformation(domain)
	if err != nil {
		return domainRecord, serverRecords, serversHasChanged, err
	}
	fmt.Printf("getOrCreateServerDomainRecords: sslinfo obtained %+v\n", sslInfo)
	domainRecord, err = getRowFromDomainsByHost(sslInfo.Host)

	if err != nil {
		if err == sql.ErrNoRows {
			logo, title, err := getLogoAndTitleFromDomain(sslInfo.Host)
			if err != nil {
				return domainRecord, serverRecords, serversHasChanged, err
			}
			domainPayload := &DatabaseDomainRow{
				Logo:  logo,
				Title: title,
				Host:  sslInfo.Host,
			}
			domainRecord, err = insertIntoDomain(domainPayload)
			if err != nil {
				return domainRecord, serverRecords, serversHasChanged, err
			}
		} else {
			if err != nil {
				return domainRecord, serverRecords, serversHasChanged, err
			}

		}
	} else {
		serverRecords, err = getRowsFromServersByDomainID(domainRecord.ID)
		if err != nil {
			return domainRecord, serverRecords, serversHasChanged, err
		}
		if len(serverRecords) > 0 {
			serversHasChanged = true
		}
	}
	for _, endpoint := range sslInfo.Endpoints {
		whoisData, err := getWhoisInfo(endpoint.IPAddress)
		if err != nil {
			return domainRecord, serverRecords, serversHasChanged, err
		}
		var serverPayload *DatabaseServesRow
		serverPayload.Address = endpoint.IPAddress
		serverPayload.DomainID = domainRecord.ID
		serverPayload.SslGrade = endpoint.Grade
		serverPayload.Owner = whoisData.owner
		serverPayload.Country = whoisData.country
		// TODO: Esto tiene que hacer un firstOrCreate para comparar correctamente
		newRow, err := insertIntoServer(serverPayload)
		if err != nil {
			return domainRecord, serverRecords, serversHasChanged, err
		}
		serverRecords = append(serverRecords, newRow)
	}
	return domainRecord, serverRecords, serversHasChanged, err
}

func testDomain(str string) bool {
	ips, err := net.LookupHost(str)
	return err == nil && len(ips) > 0
}

// Hello
func domainCreate(ctx *fasthttp.RequestCtx) {
	var postBody DomainRequest
	var err error
	err = json.Unmarshal(ctx.PostBody(), &postBody)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	if !(testDomain(postBody.Domain)) {
		ctx.Error(fmt.Sprintf("Domain in body is not a domain %+v", postBody.Domain), 403)
		return
	}
	fmt.Printf("Dominio solicitado es %+v\n", postBody.Domain)
	db, err = boostrapAndGetDatabase()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	// TODO: Hay que hacer un ping para saber si el server está caido
	domain, servers, serversHasChanged, err := getOrCreateServerDomainRecords(postBody.Domain)
	db.Close()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	var serversResponse []ServerResponse
	for _, server := range servers {
		var item ServerResponse
		item.Address = server.Address
		item.Country = server.Country
		item.Owner = server.Owner
		item.SslGrade = server.SslGrade
		serversResponse = append(serversResponse, item)
	}
	// TODO: Esto tiene que diferenciar el  por fecha
	var serverGrade string
	var previusGrade string
	if len(servers) > 0 {
		serverGrade = servers[0].SslGrade
	}
	if len(servers) > 1 {
		previusGrade = servers[len(servers)-1].SslGrade
	}
	response := &DomainResponse{
		Servers:          serversResponse,
		ServersChanged:   serversHasChanged,
		SslGrade:         serverGrade,
		PreviousSslGrade: previusGrade,
		Logo:             domain.Logo,
		Title:            domain.Title,
		IsDown:           false,
	}
	responseJSON, _ := json.Marshal(response)

	fmt.Fprintf(ctx, string(responseJSON))
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", index)
	router.POST("/dominios", domainCreate)
	log.Fatal(fasthttp.ListenAndServe(":8081", router.Handler))
}
