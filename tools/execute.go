package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExecuteBSLTool returns the MCP tool definition for execute_bsl.
func ExecuteBSLTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "execute_bsl",
		Title: "BSL kodu icra et",
		Description: "1C bazasında ixtiyari BSL kodu icra edir. " +
			"Yazma əməliyyatları: ПолучитьОбъект(), Записать(), НачатьТранзакцию(). " +
			"Output tutmaq üçün _МассивВыводаВыполнения.Добавить(\"mesaj\") istifadə et. " +
			"Diqqət: bu endpoint tam yazma icazəsinə malikdir.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"code": {
					"type": "string",
					"description": "1C BSL kodu. Output üçün _МассивВыводаВыполнения.Добавить(\"...\") istifadə et."
				}
			},
			"required": ["code"]
		}`),
	}
}

// NewExecuteBSLHandler returns a ToolHandler for execute_bsl.
func NewExecuteBSLHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.Code == "" {
			return nil, fmt.Errorf("code is required")
		}

		body := onec.ExecuteRequest{Code: input.Code}
		var result onec.ExecuteResult
		if err := client.Post(ctx, "/execute", body, &result); err != nil {
			return nil, fmt.Errorf("executing BSL in 1C: %w", err)
		}

		var b strings.Builder
		if result.Success {
			fmt.Fprintf(&b, "✅ BSL uğurla icra edildi (%d ms)\n", result.DurationMs)
		} else {
			fmt.Fprintf(&b, "❌ BSL icrasında xəta (%d ms)\n**Xəta:** %s\n", result.DurationMs, result.Error)
		}
		if len(result.Output) > 0 {
			b.WriteString("\n**Output:**\n")
			for _, line := range result.Output {
				fmt.Fprintf(&b, "- %s\n", line)
			}
		}

		return textResult(b.String()), nil
	}
}
