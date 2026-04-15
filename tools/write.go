package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WriteObjectTool returns the MCP tool definition for write_object.
func WriteObjectTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "write_object",
		Title: "Obyekt sahəsini yenilə",
		Description: "1C справочник, sənəd və ya hesab planında sahə dəyərini yeniləyir. " +
			"Tip nümunəsi: Справочник.Сотрудники, ПланСчетов.Хозрасчетный, Документ.ПриходТоваров. " +
			"Axtarış kriteriyası: {field: 'Kod', value: '01'}.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"type": {
					"type": "string",
					"description": "1C obyekt tipi. Məs: Справочник.Сотрудники, ПланСчетов.Хозрасчетный"
				},
				"search": {
					"type": "object",
					"properties": {
						"field": {"type": "string", "description": "Axtarış sahəsi (məs: Kod, Наименование)"},
						"value": {"description": "Axtarış dəyəri"}
					},
					"required": ["field", "value"]
				},
				"fields": {
					"type": "object",
					"description": "Yenilənəcək sahə-dəyər cütləri. Məs: {\"Наименование\": \"Yeni ad\"}"
				}
			},
			"required": ["type", "search", "fields"]
		}`),
	}
}

// NewWriteObjectHandler returns a ToolHandler for write_object.
func NewWriteObjectHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input onec.WriteRequest
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.Type == "" || input.Search.Field == "" || len(input.Fields) == 0 {
			return nil, fmt.Errorf("type, search.field and fields are required")
		}

		var result onec.WriteResult
		if err := client.Post(ctx, "/write", input, &result); err != nil {
			return nil, fmt.Errorf("writing object in 1C: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("write failed: server returned success=false")
		}

		return textResult(fmt.Sprintf(
			"✅ Yazıldı: %d sahə yeniləndi (ref: %s)",
			result.UpdatedFields, result.Ref,
		)), nil
	}
}
