package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mock1CHandler simulates the 1C HTTP service endpoints.
func mock1CHandler() http.Handler {
	metadata := map[string][]string{
		"Справочники": {"Контрагенты", "Номенклатура"},
		"Документы":   {"РеализацияТоваровУслуг"},
		"Регистры":    {"КурсыВалют"},
	}

	type attribute struct {
		Name    string `json:"Имя"`
		Synonym string `json:"Синоним"`
		Type    string `json:"Тип"`
	}
	type tabularSection struct {
		Name       string      `json:"Имя"`
		Attributes []attribute `json:"Реквизиты"`
	}
	type objectMeta struct {
		Name            string           `json:"Имя"`
		Synonym         string           `json:"Синоним"`
		Attributes      []attribute      `json:"Реквизиты"`
		TabularSections []tabularSection `json:"ТабличныеЧасти"`
	}

	objects := map[string]objectMeta{
		"Document/РеализацияТоваровУслуг": {
			Name:    "РеализацияТоваровУслуг",
			Synonym: "Реализация (акты, накладные, УПД)",
			Attributes: []attribute{
				{Name: "Контрагент", Synonym: "Контрагент", Type: "СправочникСсылка.Контрагенты"},
				{Name: "СуммаДокумента", Synonym: "Сумма", Type: "Число"},
			},
			TabularSections: []tabularSection{
				{
					Name: "Товары",
					Attributes: []attribute{
						{Name: "Номенклатура", Synonym: "Номенклатура", Type: "СправочникСсылка.Номенклатура"},
						{Name: "Количество", Synonym: "Количество", Type: "Число"},
					},
				},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(metadata)
	})
	mux.HandleFunc("/object/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/object/")
		obj, ok := objects[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Object not found"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(obj)
	})
	return mux
}

// setupIntegration creates a mock 1C server and connected MCP client session.
func setupIntegration(t *testing.T) (*mcp.ClientSession, func()) {
	t.Helper()

	mock := httptest.NewServer(mock1CHandler())
	client := onec.NewClient(mock.URL, "", "")
	srv := New(client)

	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	_, err := srv.Connect(ctx, st, nil)
	if err != nil {
		mock.Close()
		t.Fatalf("server connect: %v", err)
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := mcpClient.Connect(ctx, ct, nil)
	if err != nil {
		mock.Close()
		t.Fatalf("client connect: %v", err)
	}

	cleanup := func() {
		session.Close()
		mock.Close()
	}
	return session, cleanup
}

func TestIntegration_ListTools(t *testing.T) {
	session, cleanup := setupIntegration(t)
	defer cleanup()

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}

	for _, want := range []string{"get_metadata_tree", "get_object_structure", "bsl_syntax_help"} {
		if !toolNames[want] {
			t.Errorf("expected tool %q in list, got: %v", want, toolNames)
		}
	}
}

func TestIntegration_GetMetadataTree(t *testing.T) {
	session, cleanup := setupIntegration(t)
	defer cleanup()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_metadata_tree",
	})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{"Контрагенты", "Номенклатура", "РеализацияТоваровУслуг", "КурсыВалют"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in response, got:\n%s", want, text)
		}
	}
}

func TestIntegration_GetObjectStructure(t *testing.T) {
	session, cleanup := setupIntegration(t)
	defer cleanup()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_object_structure",
		Arguments: map[string]any{
			"object_type": "Document",
			"object_name": "РеализацияТоваровУслуг",
		},
	})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{"РеализацияТоваровУслуг", "Контрагент", "СуммаДокумента", "Товары", "Номенклатура"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in response, got:\n%s", want, text)
		}
	}
}

func TestIntegration_GetObjectStructure_NotFound(t *testing.T) {
	session, cleanup := setupIntegration(t)
	defer cleanup()

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_object_structure",
		Arguments: map[string]any{
			"object_type": "Document",
			"object_name": "НесуществующийДокумент",
		},
	})
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

func TestIntegration_BSLSyntaxHelp(t *testing.T) {
	session, cleanup := setupIntegration(t)
	defer cleanup()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "bsl_syntax_help",
		Arguments: map[string]any{
			"query": "СтрНайти",
		},
	})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "СтрНайти") {
		t.Errorf("expected СтрНайти in response, got:\n%s", text)
	}
	if !strings.Contains(text, "StrFind") {
		t.Errorf("expected StrFind in response, got:\n%s", text)
	}
}
