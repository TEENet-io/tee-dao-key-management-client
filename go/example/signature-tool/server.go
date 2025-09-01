// -----------------------------------------------------------------------------
// Copyright (c) 2025 TEENet Technology (Hong Kong) Limited. All Rights Reserved.
//
// This software and its associated documentation files (the "Software") are
// the proprietary and confidential information of TEENet Technology (Hong Kong) Limited.
// Unauthorized copying of this file, via any medium, is strictly prohibited.
//
// No license, express or implied, is hereby granted, except by written agreement
// with TEENet Technology (Hong Kong) Limited. Use of this software without permission
// is a violation of applicable laws.
//
// -----------------------------------------------------------------------------

package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Static file serving functionality
func staticFileHandler(frontendPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// For API routes, don't handle here
		if strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		// Handle container paths - strip container prefix if present
		if strings.HasPrefix(path, "/container/") {
			// Remove "/container/{app_id}/" prefix to get actual file path
			parts := strings.Split(strings.TrimPrefix(path, "/container/"), "/")
			if len(parts) > 1 {
				// Rebuild path without the app_id part
				path = "/" + strings.Join(parts[1:], "/")
			} else {
				path = "/"
			}
		}

		// Default to index.html for root requests
		if path == "/" {
			path = "/index.html"
		}

		// For SPA routing, we need to handle URL paths properly
		// Remove leading slash for file system path joining
		relativePath := strings.TrimPrefix(path, "/")
		
		// Security: prevent directory traversal (check for .. components)
		if strings.Contains(relativePath, "..") {
			c.String(http.StatusBadRequest, "Invalid path")
			c.Abort()
			return
		}

		filePath := filepath.Join(frontendPath, relativePath)
		
		// Determine content type based on original request path
		ext := filepath.Ext(path)
		isSPARoute := false
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Only serve index.html for non-asset requests (SPA routing)
			if ext != ".css" && ext != ".js" && ext != ".png" && ext != ".jpg" && ext != ".gif" && ext != ".ico" {
				filePath = filepath.Join(frontendPath, "index.html")
				isSPARoute = true
			} else {
				// Asset file not found, return 404
				c.String(http.StatusNotFound, "File not found")
				c.Abort()
				return
			}
		}

		// Set no-cache headers to prevent browser caching
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// Set content type
		if isSPARoute {
			c.Header("Content-Type", "text/html")
		} else {
			switch ext {
			case ".html":
				c.Header("Content-Type", "text/html")
			case ".css":
				c.Header("Content-Type", "text/css")
			case ".js":
				c.Header("Content-Type", "application/javascript")
			case ".png":
				c.Header("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				c.Header("Content-Type", "image/jpeg")
			case ".gif":
				c.Header("Content-Type", "image/gif")
			case ".ico":
				c.Header("Content-Type", "image/x-icon")
			default:
				c.Header("Content-Type", "text/plain")
			}
		}

		// Serve files directly like voting-sign-tool (no template processing)
		c.File(filePath)
		c.Abort()
	}
}