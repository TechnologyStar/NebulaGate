package main

import (
    "crypto/tls"
    "encoding/json"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "testing"
    "time"

    "github.com/QuantumNous/new-api/setting/heimdall"
)

// TestConfig holds test configuration
type TestConfig struct {
    ServerURL    string
    BackendURL   string
    TLSCertPath  string
    TLSKeyPath   string
    TestDataDir  string
}

// TestResponse represents a basic API response
type TestResponse struct {
    Message   string                 `json:"message,omitempty"`
    ProxyTarget string               `json:"proxy_target,omitempty"`
    Method    string                 `json:"method,omitempty"`
    Path      string                 `json:"path,omitempty"`
    Query     string                 `json:"query,omitempty"`
    Note      string                 `json:"note,omitempty"`
    Service   string                 `json:"service,omitempty"`
    Version   string                 `json:"version,omitempty"`
    Status    string                 `json:"status,omitempty"`
    Error     map[string]interface{} `json:"error,omitempty"`
}

func setupTestConfig(t *testing.T) *TestConfig {
    // Create temporary directory for test certificates
    testDataDir, err := os.MkdirTemp("", "heimdall-test")
    if err != nil {
        t.Fatalf("Failed to create test data directory: %v", err)
    }

    // Generate self-signed certificate for testing
    certPath := filepath.Join(testDataDir, "cert.pem")
    keyPath := filepath.Join(testDataDir, "key.pem")
    
    err = generateSelfSignedCert(certPath, keyPath)
    if err != nil {
        t.Fatalf("Failed to generate test certificate: %v", err)
    }

    return &TestConfig{
        ServerURL:   "https://localhost:8443",
        BackendURL:  "http://localhost:3000", // Mock backend
        TLSCertPath: certPath,
        TLSKeyPath:  keyPath,
        TestDataDir: testDataDir,
    }
}

func cleanupTestConfig(config *TestConfig) {
    os.RemoveAll(config.TestDataDir)
}

func generateSelfSignedCert(certPath, keyPath string) error {
    // This is a placeholder - in a real implementation, you would
    // generate actual self-signed certificates here
    // For now, we'll create empty files to simulate the structure
    
    // Create empty cert and key files
    certFile, err := os.Create(certPath)
    if err != nil {
        return err
    }
    certFile.Close()
    
    keyFile, err := os.Create(keyPath)
    if err != nil {
        return err
    }
    keyFile.Close()
    
    return nil
}

func createHTTPClient(t *testing.T, skipTLS bool) *http.Client {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: skipTLS,
        },
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second,
    }
}

func TestHeimdallService(t *testing.T) {
    config := setupTestConfig(t)
    defer cleanupTestConfig(config)

    // Set up test environment variables
    os.Setenv("HEIMDALL_TLS_ENABLED", "true")
    os.Setenv("HEIMDALL_TLS_CERT", config.TLSCertPath)
    os.Setenv("HEIMDALL_TLS_KEY", config.TLSKeyPath)
    os.Setenv("HEIMDALL_LISTEN_ADDR", ":8443")
    os.Setenv("HEIMDALL_BACKEND_URL", config.BackendURL)
    os.Setenv("HEIMDALL_LOG_LEVEL", "debug")

    // Note: In a real test environment, you would start the Heimdall service
    // in a separate goroutine and wait for it to be ready
    // For this example, we'll test the configuration loading and client creation

    t.Run("TestConfigurationLoading", func(t *testing.T) {
        testConfigLoading(t)
    })

    t.Run("TestHTTPClientCreation", func(t *testing.T) {
        testHTTPClientCreation(t, config)
    })

    t.Run("TestAPIStructure", func(t *testing.T) {
        testAPIStructure(t)
    })
}

func testConfigLoading(t *testing.T) {
    config, err := heimdall.LoadConfig()
    if err != nil {
        t.Fatalf("Failed to load configuration: %v", err)
    }

    // Test basic configuration values
    if config.ListenAddr != ":8443" {
        t.Errorf("Expected ListenAddr ':8443', got '%s'", config.ListenAddr)
    }

    if !config.TLSEnabled {
        t.Error("Expected TLSEnabled to be true")
    }

    if config.BackendURL != "http://localhost:3000" {
        t.Errorf("Expected BackendURL 'http://localhost:3000', got '%s'", config.BackendURL)
    }

    if config.LogLevel != "debug" {
        t.Errorf("Expected LogLevel 'debug', got '%s'", config.LogLevel)
    }
}

func testHTTPClientCreation(t *testing.T, config *TestConfig) {
    // Test creating HTTP client with TLS skip
    client := createHTTPClient(t, true)
    if client == nil {
        t.Error("Failed to create HTTP client")
    }

    // Test creating HTTP client without TLS skip
    secureClient := createHTTPClient(t, false)
    if secureClient == nil {
        t.Error("Failed to create secure HTTP client")
    }
}

func testAPIStructure(t *testing.T) {
    // Test that the proxy handler can be created with valid config
    config, err := heimdall.LoadConfig()
    if err != nil {
        t.Fatalf("Failed to load configuration: %v", err)
    }

    // Create proxy handler
    proxy := NewProxyHandler(config)
    if proxy == nil {
        t.Error("Failed to create proxy handler")
    }

    if proxy.config != config {
        t.Error("Proxy handler config not set correctly")
    }

    if proxy.client == nil {
        t.Error("Proxy handler HTTP client not initialized")
    }
}

func TestProxyHandler(t *testing.T) {
    config, err := heimdall.LoadConfig()
    if err != nil {
        t.Fatalf("Failed to load configuration: %v", err)
    }

    proxy := NewProxyHandler(config)

    t.Run("TestProxyHandlerCreation", func(t *testing.T) {
        if proxy == nil {
            t.Error("Failed to create proxy handler")
        }
    })

    t.Run("TestHeaderCopying", func(t *testing.T) {
        testHeaderCopying(t, proxy)
    })
}

func testHeaderCopying(t *testing.T, proxy *ProxyHandler) {
    // Test request header copying
    srcHeaders := make(http.Header)
    srcHeaders.Set("Content-Type", "application/json")
    srcHeaders.Set("Authorization", "Bearer token123")
    srcHeaders.Set("Connection", "close") // Should be filtered out
    srcHeaders.Set("Host", "example.com") // Should be filtered out for requests

    dstHeaders := make(http.Header)
    proxy.copyHeaders(srcHeaders, dstHeaders, true)

    // Check that headers were copied correctly
    if dstHeaders.Get("Content-Type") != "application/json" {
        t.Error("Content-Type header not copied correctly")
    }

    if dstHeaders.Get("Authorization") != "Bearer token123" {
        t.Error("Authorization header not copied correctly")
    }

    // Check that hop-by-hop headers were filtered
    if dstHeaders.Get("Connection") != "" {
        t.Error("Connection header should have been filtered out")
    }

    if dstHeaders.Get("Host") != "" {
        t.Error("Host header should have been filtered out for requests")
    }
}

func TestConfigurationValidation(t *testing.T) {
    t.Run("TestValidTLSConfig", func(t *testing.T) {
        // Set valid TLS configuration
        os.Setenv("HEIMDALL_TLS_ENABLED", "true")
        os.Setenv("HEIMDALL_TLS_CERT", "/tmp/test-cert.pem")
        os.Setenv("HEIMDALL_TLS_KEY", "/tmp/test-key.pem")
        os.Setenv("HEIMDALL_BACKEND_URL", "http://localhost:3000")

        // Create dummy cert files for validation
        os.Create("/tmp/test-cert.pem")
        os.Create("/tmp/test-key.pem")
        defer os.Remove("/tmp/test-cert.pem")
        defer os.Remove("/tmp/test-key.pem")

        config, err := heimdall.LoadConfig()
        if err != nil {
            t.Errorf("Valid TLS config should not fail: %v", err)
        }

        if config == nil {
            t.Error("Config should not be nil for valid configuration")
        }
    })

    t.Run("TestInvalidTLSConfig", func(t *testing.T) {
        // Set invalid TLS configuration (missing cert/key)
        os.Setenv("HEIMDALL_TLS_ENABLED", "true")
        os.Setenv("HEIMDALL_TLS_CERT", "")
        os.Setenv("HEIMDALL_TLS_KEY", "")
        os.Setenv("HEIMDALL_BACKEND_URL", "http://localhost:3000")

        config, err := heimdall.LoadConfig()
        if err == nil {
            t.Error("Invalid TLS config should fail validation")
        }

        if config != nil {
            t.Error("Config should be nil for invalid configuration")
        }
    })

    t.Run("TestACMEConfig", func(t *testing.T) {
        // Set valid ACME configuration
        os.Setenv("HEIMDALL_TLS_ENABLED", "true")
        os.Setenv("HEIMDALL_ACME_ENABLED", "true")
        os.Setenv("HEIMDALL_ACME_DOMAIN", "example.com")
        os.Setenv("HEIMDALL_ACME_EMAIL", "admin@example.com")
        os.Setenv("HEIMDALL_BACKEND_URL", "http://localhost:3000")

        config, err := heimdall.LoadConfig()
        if err != nil {
            t.Errorf("Valid ACME config should not fail: %v", err)
        }

        if !config.ACMEEnabled {
            t.Error("ACME should be enabled")
        }

        if config.ACMEDomain != "example.com" {
            t.Errorf("Expected domain 'example.com', got '%s'", config.ACMEDomain)
        }
    })
}

func TestEnvironmentParsing(t *testing.T) {
    t.Run("TestBoolParsing", func(t *testing.T) {
        testCases := []struct {
            value    string
            expected bool
        }{
            {"true", true},
            {"false", false},
            {"TRUE", true},
            {"FALSE", false},
            {"", false}, // default value
            {"invalid", false}, // default value on error
        }

        for _, tc := range testCases {
            os.Setenv("TEST_BOOL", tc.value)
            result := getEnvBoolWithDefault("TEST_BOOL", false)
            if result != tc.expected {
                t.Errorf("Expected %v for value '%s', got %v", tc.expected, tc.value, result)
            }
        }
    })

    t.Run("TestIntParsing", func(t *testing.T) {
        testCases := []struct {
            value    string
            expected int
        }{
            {"123", 123},
            {"0", 0},
            {"-1", -1},
            {"", 42}, // default value
            {"invalid", 42}, // default value on error
        }

        for _, tc := range testCases {
            os.Setenv("TEST_INT", tc.value)
            result := getEnvIntWithDefault("TEST_INT", 42)
            if result != tc.expected {
                t.Errorf("Expected %d for value '%s', got %d", tc.expected, tc.value, result)
            }
        }
    })
}

// Benchmark tests
func BenchmarkProxyHandlerCreation(b *testing.B) {
    config, _ := heimdall.LoadConfig()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        proxy := NewProxyHandler(config)
        _ = proxy
    }
}

func BenchmarkHeaderCopying(b *testing.B) {
    config, _ := heimdall.LoadConfig()
    proxy := NewProxyHandler(config)
    
    srcHeaders := make(http.Header)
    srcHeaders.Set("Content-Type", "application/json")
    srcHeaders.Set("Authorization", "Bearer token123")
    srcHeaders.Set("X-Custom-Header", "custom-value")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dstHeaders := make(http.Header)
        proxy.copyHeaders(srcHeaders, dstHeaders, true)
    }
}

// Helper functions for testing
func getEnvBoolWithDefault(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.ParseBool(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}

func getEnvIntWithDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.Atoi(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}

// Integration test example (requires actual service running)
func TestHeimdallIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // This test would require the Heimdall service to be running
    // In a CI/CD environment, you would start the service before running tests
    
    t.Run("TestServiceHealth", func(t *testing.T) {
        client := createHTTPClient(t, true) // Skip TLS verification for testing
        
        resp, err := client.Get("https://localhost:8443/health")
        if err != nil {
            t.Skipf("Heimdall service not running: %v", err)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            t.Errorf("Expected status 200, got %d", resp.StatusCode)
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            t.Fatalf("Failed to read response body: %v", err)
        }

        var healthResp TestResponse
        if err := json.Unmarshal(body, &healthResp); err != nil {
            t.Fatalf("Failed to parse JSON response: %v", err)
        }

        if healthResp.Service != "heimdall" {
            t.Errorf("Expected service 'heimdall', got '%s'", healthResp.Service)
        }
    })

    t.Run("TestRootEndpoint", func(t *testing.T) {
        client := createHTTPClient(t, true)
        
        resp, err := client.Get("https://localhost:8443/")
        if err != nil {
            t.Skipf("Heimdall service not running: %v", err)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            t.Errorf("Expected status 200, got %d", resp.StatusCode)
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            t.Fatalf("Failed to read response body: %v", err)
        }

        var rootResp TestResponse
        if err := json.Unmarshal(body, &rootResp); err != nil {
            t.Fatalf("Failed to parse JSON response: %v", err)
        }

        if rootResp.Service != "Heimdall Gateway" {
            t.Errorf("Expected service 'Heimdall Gateway', got '%s'", rootResp.Service)
        }
    })
}