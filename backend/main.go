package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Define structs to parse API responses
type ANNOqResponse struct {
	GeneID   string `json:"gene_id"`
	Annotation string `json:"annotation"`
	// Add other fields as needed
}

type PANTHERResponse struct {
	GeneID         string `json:"gene_id"`
	AdditionalInfo string `json:"additional_info"`
	// Add other fields as needed
}

func main() {
	router := gin.Default()

	// Enable CORS for frontend-backend communication
	router.Use(cors.Default())

	// Endpoint for file upload
	router.POST("/upload", uploadHandler)

	// Endpoint to query databases and get combined results
	router.GET("/query", queryHandler)

	// Start server on port 8080
	router.Run(":8080")
}

func uploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}

	// Open the uploaded file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open file"})
		return
	}
	defer f.Close()

	// Read the CSV file
	r := csv.NewReader(f)
	genes := []string{}

	// Skip the header if present
	_, err = r.Read()

	// Read each row of the CSV
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CSV file"})
			return
		}

		genes = append(genes, record[0]) // Assuming gene ID is in the first column
	}

	// Store the list of genes in the context for later querying
	c.Set("genes", genes)

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "genes": genes})
}

func queryHandler(c *gin.Context) {
	// Retrieve the list of genes from the context
	genes, ok := c.Get("genes")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No genes found in context"})
		return
	}

	geneList := genes.([]string)
	results := []map[string]string{}

	for _, gene := range geneList {
		// Query ANNOq API
		annoqData, err := queryANNOqAPI(gene)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query ANNOq API"})
			return
		}

		// Query PANTHER API
		pantherData, err := queryPANTHERAPI(annoqData.GeneID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query PANTHER API"})
			return
		}

		// Combine the results from both APIs
		result := map[string]string{
			"GeneID":         annoqData.GeneID,
			"Annotation":     annoqData.Annotation,
			"AdditionalInfo": pantherData.AdditionalInfo,
		}

		results = append(results, result)
	}

	// Return the combined results as JSON
	c.JSON(http.StatusOK, results)
}

func queryANNOqAPI(geneID string) (ANNOqResponse, error) {
	// Replace with the actual ANNOq API endpoint
	apiURL := fmt.Sprintf("http://annoq.org/api/query/%s", geneID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return ANNOqResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ANNOqResponse{}, fmt.Errorf("Failed to fetch data from ANNOq API")
	}

	var annoqData ANNOqResponse
	if err := json.NewDecoder(resp.Body).Decode(&annoqData); err != nil {
		return ANNOqResponse{}, err
	}

	return annoqData, nil
}

func queryPANTHERAPI(geneID string) (PANTHERResponse, error) {
	// Replace with the actual PANTHER API endpoint
	apiURL := fmt.Sprintf("http://pantherdb.org/api/query/%s", geneID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return PANTHERResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PANTHERResponse{}, fmt.Errorf("Failed to fetch data from PANTHER API")
	}

	var pantherData PANTHERResponse
	if err := json.NewDecoder(resp.Body).Decode(&pantherData); err != nil {
		return PANTHERResponse{}, err
	}

	return pantherData, nil
}
