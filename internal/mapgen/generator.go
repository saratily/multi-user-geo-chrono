// Package mapgen provides HTML map generation functionality for GPS track visualization.
//
// @title HTML Map Generator Package
// @version 1.0
// @description Creates interactive Google Maps for GPS tracking data visualization
// @description Generates HTML files with markers, paths, and information windows
// @description Supports extensive customization through configuration options
//
// Features:
// - Interactive Google Maps integration
// - Customizable markers and paths
// - Information windows with GPS data
// - Responsive web design
// - Template-based HTML generation
package mapgen

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/saratily/geo-chrono/internal/config"
	"github.com/saratily/geo-chrono/internal/gps"
)

// Generator handles the creation of HTML files containing interactive Google Maps
// for GPS track visualization.
//
// @struct Generator
// @description HTML map generator with Google Maps integration
// @description Uses Go templates to create dynamic web pages with JavaScript
// @property config Config Configuration settings for map appearance and behavior
type Generator struct {
	config *config.Config // @field config Configuration settings for map appearance and behavior
}

// NewGenerator creates a new map generator instance with the provided configuration.
//
// @function NewGenerator
// @description Creates configured HTML map generator
// @param cfg *config.Config Map styling, controls, and display configuration
// @return *Generator Configured map generator instance
// @example generator := NewGenerator(config)
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{config: cfg}
}

// MapData holds all the data required for HTML template execution and map generation.
//
// @struct MapData
// @description Template context object for HTML map generation
// @description Contains GPS data, credentials, and configuration for rendering
// @property Points gps.Points GPS tracking points to visualize on map
// @property APIKey string Google Maps API authentication key
// @property Title string HTML page title and header text
// @property OutputFile string Target file path for generated HTML
// @property Config Config Complete configuration for template access
type MapData struct {
	Points     gps.Points     // @field Points GPS points to display on the map
	APIKey     string         // @field APIKey Google Maps API key for map service authentication
	Title      string         // @field Title Title to display at the top of the generated HTML page
	OutputFile string         // @field OutputFile Target file path for the generated HTML output
	Config     *config.Config // @field Config Complete configuration object for template access
}

// Generate creates a complete HTML file containing an interactive Google Map visualization
// of the provided GPS points.
//
// @method Generate
// @description Creates interactive HTML map file from GPS tracking data
// @param points gps.Points Collection of GPS points to visualize
// @param outputFile string Target file path for generated HTML
// @return error Error if template processing or file creation fails
// @output HTML file with Google Maps, markers, paths, and info windows
// @browser Compatible with modern web browsers, requires internet connection
// @example err := generator.Generate(gpsPoints, "map.html")
func (g *Generator) Generate(points gps.Points, outputFile string) error {
	// Prepare all data needed for template execution
	mapData := MapData{
		Points:     points,                     // GPS tracking points to visualize
		APIKey:     g.config.GoogleMaps.APIKey, // Authentication for Google Maps API
		Title:      g.config.Map.Title,         // Page title from configuration
		OutputFile: outputFile,                 // Target file path for HTML output
		Config:     g.config,                   // Full config for template access
	}

	// Generate the HTML file using the prepared data
	return g.generateHTML(mapData)
}

// generateHTML creates the HTML file with embedded Google Maps functionality.
//
// @method generateHTML
// @description Processes HTML template and writes final map file
// @param data MapData Template context with GPS data and configuration
// @return error Error if template processing or file writing fails
// @internal true
// @steps Parse template, Register functions, Create file, Execute template
func (g *Generator) generateHTML(data MapData) error {
	// Get the HTML template containing the complete page structure
	tmpl := g.getHTMLTemplate()

	// Define custom template functions for use within the HTML template
	// These functions provide additional formatting and utility capabilities
	funcMap := template.FuncMap{
		"add":   func(a, b int) int { return a + b },                                         // Mathematical addition for indexing
		"sub":   func(a, b int) int { return a - b },                                         // Mathematical subtraction
		"upper": func(s string) string { return strings.ToUpper(s) },                         // String case conversion
		"join":  func(slice []string, sep string) string { return strings.Join(slice, sep) }, // Array joining for parameters
	}

	// Parse the template with custom functions registered
	t, err := template.New("map").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	// Create the output HTML file
	file, err := os.Create(data.OutputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	// Execute the template with the map data, generating the final HTML content
	err = t.Execute(file, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

// getHTMLTemplate returns the complete HTML template for GPS track visualization.
//
// @method getHTMLTemplate
// @description Returns complete HTML template for GPS map visualization
// @return string Full HTML template with embedded CSS and JavaScript
// @internal true
// @components Responsive CSS, Statistics display, Google Maps, Legend, JavaScript
// @features Dynamic content, Custom markers, Path drawing, Info windows
// @template Integrated with Go template system for data binding
func (g *Generator) getHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .header h1 {
            color: #333;
            margin: 0;
        }
        .stats {
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
            text-align: center;
        }
        .stats span {
            display: inline-block;
            margin: 0 20px;
            color: #666;
        }
        #map {
            height: {{.Config.Map.Height}};
            width: {{.Config.Map.Width}};
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .legend {
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-top: 20px;
        }
        .legend h3 {
            margin-top: 0;
            color: #333;
        }
        .legend-item {
            display: inline-block;
            margin: 5px 15px 5px 0;
        }
        .legend-color {
            display: inline-block;
            width: 20px;
            height: 20px;
            margin-right: 8px;
            vertical-align: middle;
            border-radius: 50%;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Title}}</h1>
    </div>

    {{if .Points}}
    <div class="stats">
        <span><strong>Total Points:</strong> {{len .Points}}</span>
        <span><strong>Start:</strong> {{(.Points.First).Timestamp.Format "2006-01-02 15:04"}}</span>
        <span><strong>End:</strong> {{(.Points.Last).Timestamp.Format "2006-01-02 15:04"}}</span>
    </div>
    {{end}}

    <div id="map"></div>

    <div class="legend">
        <h3>Legend</h3>
        <div class="legend-item">
            <span class="legend-color" style="background-color: #00FF00;"></span>
            Start Point
        </div>
        <div class="legend-item">
            <span class="legend-color" style="background-color: #FF0000;"></span>
            End Point
        </div>
        <div class="legend-item">
            <span class="legend-color" style="background-color: #0000FF;"></span>
            Waypoints
        </div>
        <div class="legend-item">
            <span style="display: inline-block; width: 30px; height: 3px; background-color: {{.Config.Path.Style.Color}}; margin-right: 8px; vertical-align: middle;"></span>
            Walking Trail
        </div>
    </div>

    <script>
        let map;

        const points = [
            {{range $i, $point := .Points}}
            {
                lat: {{$point.Latitude}},
                lng: {{$point.Longitude}},
                timestamp: "{{$point.Timestamp.Format "2006-01-02 15:04:05"}}",
                title: "{{if $point.Title}}{{$point.Title}}{{else}}Point {{add $i 1}}{{end}}",
                description: "{{$point.Description}}",
                index: {{$i}}
            },
            {{end}}
        ];

        function initMap() {
            if (points.length === 0) {
                document.getElementById('map').innerHTML = '<div style="text-align: center; padding: 50px; color: #666;">No GPS points to display</div>';
                return;
            }

            // Initialize map
            const center = {{if and .Config.Map.InitialView.Center.Latitude .Config.Map.InitialView.Center.Longitude}}{lat: {{.Config.Map.InitialView.Center.Latitude}}, lng: {{.Config.Map.InitialView.Center.Longitude}}}{{else}}calculateCenter(points){{end}};
            
            map = new google.maps.Map(document.getElementById("map"), {
                zoom: {{if .Config.Map.InitialView.Zoom}}{{.Config.Map.InitialView.Zoom}}{{else}}13{{end}},
                center: center,
                mapTypeId: google.maps.MapTypeId.ROADMAP,
                zoomControl: {{.Config.Map.Controls.ZoomControl}},
                streetViewControl: {{.Config.Map.Controls.StreetViewControl}},
                fullscreenControl: {{.Config.Map.Controls.FullscreenControl}},
                mapTypeControl: {{.Config.Map.Controls.MapTypeControl}},
                scaleControl: {{.Config.Map.Controls.ScaleControl}}
            });

            // Add markers
            addMarkers();
            
            // Add walking path
            {{if .Config.Path.Enabled}}
            addWalkingPath();
            {{end}}
            
            // Fit map to show all points
            fitMapToBounds();
        }

        function calculateCenter(points) {
            let lat = 0, lng = 0;
            points.forEach(point => {
                lat += point.lat;
                lng += point.lng;
            });
            return {
                lat: lat / points.length,
                lng: lng / points.length
            };
        }

        function addMarkers() {
            points.forEach((point, index) => {
                let icon, title = point.title;

                // Customize marker icons
                if (index === 0) {
                    icon = createMarkerIcon('#00FF00', 'S', 32);
                    title = "START - " + title;
                } else if (index === points.length - 1) {
                    icon = createMarkerIcon('#FF0000', 'E', 32);
                    title = "END - " + title;
                } else {
                    icon = createMarkerIcon('#0000FF', (index + 1).toString(), 24);
                }

                const marker = new google.maps.Marker({
                    position: { lat: point.lat, lng: point.lng },
                    map: map,
                    title: title,
                    icon: icon
                });

                // Info window
                {{if .Config.InfoWindows.Enabled}}
                const infoWindow = new google.maps.InfoWindow({
                    content: createInfoWindowContent(point, title, index),
                    maxWidth: {{.Config.InfoWindows.MaxWidth}}
                });

                marker.addListener("click", () => {
                    infoWindow.open(map, marker);
                });
                {{end}}
            });
        }

        function createMarkerIcon(color, text, size) {
            return {
                url: 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(
                    '<svg xmlns="http://www.w3.org/2000/svg" width="' + size + '" height="' + size + '" viewBox="0 0 ' + size + ' ' + size + '">' +
                    '<circle cx="' + (size/2) + '" cy="' + (size/2) + '" r="' + (size/2-2) + '" fill="' + color + '" stroke="#000" stroke-width="2"/>' +
                    '<text x="' + (size/2) + '" y="' + (size/2+4) + '" text-anchor="middle" fill="white" font-family="Arial" font-size="' + (size/3) + '" font-weight="bold">' + text + '</text>' +
                    '</svg>'
                ),
                scaledSize: new google.maps.Size(size, size),
                anchor: new google.maps.Point(size/2, size/2)
            };
        }

        function createInfoWindowContent(point, title, index) {
            return ` + "`" + `
                <div style="font-family: Arial, sans-serif; min-width: 200px;">
                    <h3 style="margin: 0 0 10px 0; color: #333;">${title}</h3>
                    <p><strong>Time:</strong> ${point.timestamp}</p>
                    <p><strong>Location:</strong> ${point.lat.toFixed(6)}, ${point.lng.toFixed(6)}</p>
                    <p><strong>Sequence:</strong> ${index + 1} of ${points.length}</p>
                    ${point.description ? '<p><strong>Description:</strong> ' + point.description + '</p>' : ''}
                </div>
            ` + "`" + `;
        }

        function addWalkingPath() {
            const pathCoordinates = points.map(point => ({ lat: point.lat, lng: point.lng }));

            const walkingPath = new google.maps.Polyline({
                path: pathCoordinates,
                geodesic: true,
                strokeColor: "{{.Config.Path.Style.Color}}",
                strokeOpacity: {{.Config.Path.Style.Opacity}},
                strokeWeight: {{.Config.Path.Style.Weight}},
            });

            walkingPath.setMap(map);

            // Add direction arrows
            {{if .Config.Path.Animation.ShowDirectionArrows}}
            const arrowSymbol = {
                path: google.maps.SymbolPath.FORWARD_CLOSED_ARROW,
                scale: 3,
                strokeColor: "{{.Config.Path.Style.Color}}",
                fillColor: "{{.Config.Path.Style.Color}}",
                fillOpacity: 1
            };

            const arrowPath = new google.maps.Polyline({
                path: pathCoordinates,
                geodesic: true,
                strokeOpacity: 0,
                icons: [{
                    icon: arrowSymbol,
                    offset: '100%',
                    repeat: '100px'
                }],
            });

            arrowPath.setMap(map);
            {{end}}
        }

        function fitMapToBounds() {
            {{if .Config.Map.AutoFitBounds}}
            const bounds = new google.maps.LatLngBounds();
            points.forEach(point => {
                bounds.extend({ lat: point.lat, lng: point.lng });
            });
            map.fitBounds(bounds);
            
            // Ensure minimum zoom level
            google.maps.event.addListenerOnce(map, 'bounds_changed', function() {
                if (map.getZoom() > 15) {
                    map.setZoom(15);
                }
            });
            {{end}}
        }

        // Helper function for template
        window.initMap = initMap;
    </script>
    <script async defer src="https://maps.googleapis.com/maps/api/js?key={{.APIKey}}&callback=initMap{{if .Config.GoogleMaps.Libraries}}&libraries={{join .Config.GoogleMaps.Libraries ","}}{{end}}"></script>
</body>
</html>`
}
