package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type globalOpts struct {
	serverCmd string
	serverURL string
	timeout   time.Duration
	retries   int
	jsonOnly  bool
}

type mcpClient interface {
	Initialize(ctx context.Context) error
	Call(ctx context.Context, method string, params any, out any) error
	Close() error
}

type stdioClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bytes.Buffer
	nextID int64
}

type httpClient struct {
	url    string
	client *http.Client
	nextID int64
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpToolsList struct {
	Tools []struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"inputSchema"`
	} `json:"tools"`
}

type mcpToolCallResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

type contractPayload struct {
	Data    map[string]any `json:"data"`
	AsOf    string         `json:"as_of"`
	Quality string         `json:"quality"`
	Source  []string       `json:"source"`
	Errors  []struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details"`
	} `json:"errors"`
}

func main() {
	opts, cmd, args, err := parseGlobal(os.Args[1:])
	if err != nil {
		exitErr(err)
	}
	if cmd == "" || cmd == "help" || cmd == "--help" || cmd == "-h" {
		printHelp()
		return
	}
	if cmd == "setup" {
		exitIfErr(runSetup(args))
		return
	}
	if cmd == "doctor" {
		exitIfErr(runDoctor(opts))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()

	client, err := newClient(ctx, opts)
	if err != nil {
		exitErr(err)
	}
	defer client.Close()

	if err := client.Initialize(ctx); err != nil {
		exitErr(fmt.Errorf("initialize MCP failed: %w", err))
	}

	switch cmd {
	case "tools":
		exitIfErr(runTools(ctx, client))
	case "call":
		exitIfErr(runCall(ctx, client, opts, args))
	case "company":
		exitIfErr(runCompany(ctx, client, opts, args))
	case "chain":
		exitIfErr(runChain(ctx, client, opts, args))
	case "bottleneck":
		exitIfErr(runBottleneck(ctx, client, opts, args))
	case "macro":
		exitIfErr(runMacro(ctx, client, opts, args))
	case "search":
		exitIfErr(runSearchCompanies(ctx, client, opts, args))
	case "semantic":
		exitIfErr(runSemantic(ctx, client, opts, args))
	case "earnings":
		exitIfErr(runEarnings(ctx, client, opts, args))
	case "reports":
		exitIfErr(runReports(ctx, client, opts, args))
	case "options":
		exitIfErr(runOptions(ctx, client, opts, args))
	case "etf":
		exitIfErr(runETF(ctx, client, opts, args))
	case "technicals":
		exitIfErr(runTechnicals(ctx, client, opts, args))
	default:
		exitErr(fmt.Errorf("unknown command: %s", cmd))
	}
}

func parseGlobal(args []string) (globalOpts, string, []string, error) {
	cfg := loadFileConfig()
	opts := globalOpts{
		serverCmd: firstNonEmpty(strings.TrimSpace(os.Getenv("RW_MCP_SERVER_CMD")), cfg["RW_MCP_SERVER_CMD"]),
		serverURL: firstNonEmpty(strings.TrimSpace(os.Getenv("RW_MCP_SERVER_URL")), cfg["RW_MCP_SERVER_URL"]),
		timeout:   30 * time.Second,
		retries:   1,
		jsonOnly:  false,
	}

	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--server-cmd":
			if i+1 >= len(args) {
				return opts, "", nil, errors.New("missing value for --server-cmd")
			}
			opts.serverCmd = strings.TrimSpace(args[i+1])
			i++
		case "--server-url":
			if i+1 >= len(args) {
				return opts, "", nil, errors.New("missing value for --server-url")
			}
			opts.serverURL = strings.TrimSpace(args[i+1])
			i++
		case "--timeout":
			if i+1 >= len(args) {
				return opts, "", nil, errors.New("missing value for --timeout")
			}
			sec, err := strconv.Atoi(args[i+1])
			if err != nil || sec <= 0 {
				return opts, "", nil, errors.New("--timeout must be positive integer seconds")
			}
			opts.timeout = time.Duration(sec) * time.Second
			i++
		case "--retries":
			if i+1 >= len(args) {
				return opts, "", nil, errors.New("missing value for --retries")
			}
			r, err := strconv.Atoi(args[i+1])
			if err != nil || r < 0 {
				return opts, "", nil, errors.New("--retries must be >= 0")
			}
			opts.retries = r
			i++
		case "--json":
			opts.jsonOnly = true
		default:
			rest = append(rest, a)
		}
	}

	if len(rest) == 0 {
		return opts, "", nil, nil
	}
	if rest[0] == "help" || rest[0] == "--help" || rest[0] == "-h" || rest[0] == "setup" || rest[0] == "doctor" {
		return opts, rest[0], rest[1:], nil
	}
	if opts.serverCmd == "" && opts.serverURL == "" {
		return opts, "", nil, errors.New("missing MCP server config, set --server-url or --server-cmd (or RW_MCP_SERVER_URL/RW_MCP_SERVER_CMD)")
	}
	if opts.serverCmd != "" && opts.serverURL != "" {
		return opts, "", nil, errors.New("set only one transport: --server-url or --server-cmd")
	}
	return opts, rest[0], rest[1:], nil
}

func loadFileConfig() map[string]string {
	cfg := map[string]string{}
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	p := filepath.Join(home, ".config", "rw", "config.env")
	b, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		cfg[strings.TrimSpace(kv[0])] = strings.Trim(strings.TrimSpace(kv[1]), `"`)
	}
	return cfg
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func newClient(ctx context.Context, opts globalOpts) (mcpClient, error) {
	if opts.serverURL != "" {
		return &httpClient{
			url: strings.TrimRight(opts.serverURL, "/"),
			client: &http.Client{
				Timeout: opts.timeout,
			},
			nextID: 1,
		}, nil
	}
	return newStdioClient(ctx, opts.serverCmd)
}

func newStdioClient(ctx context.Context, serverCmd string) (*stdioClient, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", serverCmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start server failed: %w", err)
	}

	c := &stdioClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		stderr: &stderr,
		nextID: 1,
	}
	return c, nil
}

func (c *stdioClient) Close() error {
	_ = c.stdin.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	return nil
}

func (c *stdioClient) Initialize(ctx context.Context) error {
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "rw-cli",
			"version": "0.2.0",
		},
	}
	var initResult map[string]any
	if err := c.Call(ctx, "initialize", params, &initResult); err != nil {
		return err
	}
	return c.Notify(ctx, "notifications/initialized", map[string]any{})
}

func (c *stdioClient) Notify(_ context.Context, method string, params any) error {
	req := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	return c.writeMessage(req)
}

func (c *stdioClient) Call(ctx context.Context, method string, params any, out any) error {
	id := c.nextID
	c.nextID++
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	if err := c.writeMessage(req); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := c.readMessage()
		if err != nil {
			return err
		}
		if resp.ID == nil {
			continue
		}
		if *resp.ID != id {
			continue
		}
		if resp.Error != nil {
			return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(resp.Result, out)
	}
}

func (c *stdioClient) writeMessage(msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(c.stdin, header); err != nil {
		return err
	}
	_, err = c.stdin.Write(body)
	return err
}

func (c *stdioClient) readMessage() (*rpcResponse, error) {
	contentLength := 0
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && c.stderr != nil && strings.TrimSpace(c.stderr.String()) != "" {
				return nil, fmt.Errorf("server exited: %s", strings.TrimSpace(c.stderr.String()))
			}
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			v := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(line), "content-length:"))
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("bad content-length: %q", v)
			}
			contentLength = n
		}
	}
	if contentLength <= 0 {
		return nil, errors.New("missing content-length")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(c.stdout, body); err != nil {
		return nil, err
	}
	var resp rpcResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *httpClient) Initialize(ctx context.Context) error {
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "rw-cli",
			"version": "0.2.0",
		},
	}
	var initResult map[string]any
	if err := h.Call(ctx, "initialize", params, &initResult); err != nil {
		return err
	}
	// some HTTP MCP gateways ignore notifications; treat failure as non-fatal
	_ = h.notify(ctx, "notifications/initialized", map[string]any{})
	return nil
}

func (h *httpClient) Close() error { return nil }

func (h *httpClient) Call(ctx context.Context, method string, params any, out any) error {
	id := h.nextID
	h.nextID++
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	resp, err := h.post(ctx, reqBody)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

func (h *httpClient) notify(ctx context.Context, method string, params any) error {
	_, err := h.post(ctx, map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
	return err
}

func (h *httpClient) post(ctx context.Context, payload any) (*rpcResponse, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return nil, fmt.Errorf("http %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var r rpcResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func runTools(ctx context.Context, c mcpClient) error {
	var list mcpToolsList
	if err := c.Call(ctx, "tools/list", map[string]any{}, &list); err != nil {
		return err
	}
	for _, t := range list.Tools {
		fmt.Printf("%s\n", t.Name)
		if t.Description != "" {
			fmt.Printf("  %s\n", t.Description)
		}
	}
	return nil
}

func runCall(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("call", flag.ContinueOnError)
	tool := fs.String("tool", "", "MCP tool name")
	argsJSON := fs.String("args", "{}", "JSON object arguments")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *tool == "" && fs.NArg() > 0 {
		*tool = fs.Arg(0)
	}
	if *tool == "" {
		return errors.New("missing tool name")
	}
	var params map[string]any
	if err := json.Unmarshal([]byte(*argsJSON), &params); err != nil {
		return fmt.Errorf("invalid --args JSON: %w", err)
	}
	return callAndPrint(ctx, c, opts, *tool, params)
}

func runCompany(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("company", flag.ContinueOnError)
	ticker := fs.String("ticker", "", "Stock ticker")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *ticker == "" {
		return errors.New("missing --ticker")
	}
	return callAndPrint(ctx, c, opts, "query_company", map[string]any{"ticker": strings.ToUpper(strings.TrimSpace(*ticker))})
}

func runChain(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("chain", flag.ContinueOnError)
	entity := fs.String("entity", "", "Ticker or material")
	direction := fs.String("direction", "both", "upstream|downstream|both")
	maxDepth := fs.Int("max-depth", 3, "Max depth")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *entity == "" {
		return errors.New("missing --entity")
	}
	return callAndPrint(ctx, c, opts, "query_supply_chain", map[string]any{
		"entity":    strings.TrimSpace(*entity),
		"direction": strings.TrimSpace(*direction),
		"max_depth": *maxDepth,
	})
}

func runBottleneck(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("bottleneck", flag.ContinueOnError)
	domain := fs.String("domain", "", "memory|photonics|packaging|power|gpu")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *domain == "" {
		return errors.New("missing --domain")
	}
	return callAndPrint(ctx, c, opts, "query_bottleneck", map[string]any{"domain": strings.TrimSpace(*domain)})
}

func runMacro(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("macro", flag.ContinueOnError)
	days := fs.Int("days", 30, "Lookback days")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return callAndPrint(ctx, c, opts, "query_macro", map[string]any{"days": *days})
}

func runSearchCompanies(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	sector := fs.String("sector", "", "Sector filter")
	tier := fs.String("tier", "", "chokepoint|enabler|beneficiary")
	country := fs.String("country", "", "Country code")
	minCap := fs.Int64("min-cap", 0, "Min market cap")
	maxCap := fs.Int64("max-cap", 0, "Max market cap")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	params := map[string]any{}
	if *sector != "" {
		params["sector"] = *sector
	}
	if *tier != "" {
		params["tier"] = *tier
	}
	if *country != "" {
		params["country"] = *country
	}
	if *minCap > 0 {
		params["min_cap"] = *minCap
	}
	if *maxCap > 0 {
		params["max_cap"] = *maxCap
	}
	return callAndPrint(ctx, c, opts, "search_companies", params)
}

func runSemantic(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("semantic", flag.ContinueOnError)
	query := fs.String("query", "", "Natural language query")
	limit := fs.Int("limit", 5, "Top N")
	source := fs.String("source-filter", "", "Optional source")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *query == "" {
		return errors.New("missing --query")
	}
	params := map[string]any{"query": *query, "limit": *limit}
	if *source != "" {
		params["source_filter"] = *source
	}
	return callAndPrint(ctx, c, opts, "search_semantic", params)
}

func runEarnings(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("earnings", flag.ContinueOnError)
	ticker := fs.String("ticker", "", "Ticker")
	quarter := fs.String("quarter", "", "Quarter e.g. Q4-2025")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *ticker == "" {
		return errors.New("missing --ticker")
	}
	params := map[string]any{"ticker": strings.ToUpper(strings.TrimSpace(*ticker))}
	if *quarter != "" {
		params["quarter"] = strings.TrimSpace(*quarter)
	}
	return callAndPrint(ctx, c, opts, "get_earnings_call", params)
}

func runReports(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("reports", flag.ContinueOnError)
	ticker := fs.String("ticker", "", "Ticker")
	topic := fs.String("topic", "", "Topic keyword")
	source := fs.String("source", "", "Source filter")
	weeks := fs.Int("weeks", 8, "Recent weeks")
	limit := fs.Int("limit", 5, "Top N")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *ticker == "" && *topic == "" {
		return errors.New("one of --ticker or --topic is required")
	}
	params := map[string]any{"weeks": *weeks, "limit": *limit}
	if *ticker != "" {
		params["ticker"] = strings.ToUpper(strings.TrimSpace(*ticker))
	}
	if *topic != "" {
		params["topic"] = strings.TrimSpace(*topic)
	}
	if *source != "" {
		params["source"] = strings.TrimSpace(*source)
	}
	return callAndPrint(ctx, c, opts, "get_research_reports", params)
}

func runOptions(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("options", flag.ContinueOnError)
	ticker := fs.String("ticker", "", "Ticker")
	days := fs.Int("days", 30, "Lookback days")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *ticker == "" {
		return errors.New("missing --ticker")
	}
	return callAndPrint(ctx, c, opts, "query_options_flow", map[string]any{
		"ticker": strings.ToUpper(strings.TrimSpace(*ticker)),
		"days":   *days,
	})
}

func runETF(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("etf", flag.ContinueOnError)
	ticker := fs.String("ticker", "", "Optional ETF ticker")
	days := fs.Int("days", 30, "Lookback days")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	params := map[string]any{"days": *days}
	if *ticker != "" {
		params["ticker"] = strings.ToUpper(strings.TrimSpace(*ticker))
	}
	return callAndPrint(ctx, c, opts, "query_etf_flow", params)
}

func runTechnicals(ctx context.Context, c mcpClient, opts globalOpts, args []string) error {
	fs := flag.NewFlagSet("technicals", flag.ContinueOnError)
	ticker := fs.String("ticker", "", "Ticker")
	days := fs.Int("days", 7, "Lookback days")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *ticker == "" {
		return errors.New("missing --ticker")
	}
	return callAndPrint(ctx, c, opts, "query_technicals", map[string]any{
		"ticker": strings.ToUpper(strings.TrimSpace(*ticker)),
		"days":   *days,
	})
}

func callAndPrint(ctx context.Context, c mcpClient, opts globalOpts, tool string, params map[string]any) error {
	var lastErr error
	for i := 0; i <= opts.retries; i++ {
		payload, rawResult, err := callTool(ctx, c, tool, params)
		if err != nil {
			lastErr = err
		} else {
			if payload.Quality == "error" && hasErrorCode(payload, "INTERNAL_ERROR") && i < opts.retries {
				lastErr = fmt.Errorf("tool returned INTERNAL_ERROR")
			} else {
				if opts.jsonOnly {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(payload)
				}
				printSummary(tool, params, payload, rawResult)
				return nil
			}
		}
		if i < opts.retries {
			backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
			time.Sleep(backoff)
		}
	}
	return lastErr
}

func callTool(ctx context.Context, c mcpClient, tool string, params map[string]any) (contractPayload, mcpToolCallResult, error) {
	callParams := map[string]any{
		"name":      tool,
		"arguments": params,
	}
	var result mcpToolCallResult
	if err := c.Call(ctx, "tools/call", callParams, &result); err != nil {
		return contractPayload{}, mcpToolCallResult{}, err
	}
	if len(result.Content) == 0 || result.Content[0].Text == "" {
		return contractPayload{}, result, errors.New("empty content in tools/call result")
	}
	var payload contractPayload
	if err := json.Unmarshal([]byte(result.Content[0].Text), &payload); err != nil {
		return contractPayload{}, result, fmt.Errorf("failed to parse contract payload: %w", err)
	}
	return payload, result, nil
}

func printSummary(tool string, _ map[string]any, payload contractPayload, raw mcpToolCallResult) {
	fmt.Printf("tool: %s\n", tool)
	fmt.Printf("quality: %s\n", payload.Quality)
	fmt.Printf("as_of: %s\n", payload.AsOf)
	fmt.Printf("source: %s\n", strings.Join(payload.Source, ", "))
	if len(payload.Errors) == 0 {
		fmt.Println("errors: none")
	} else {
		fmt.Println("errors:")
		for _, e := range payload.Errors {
			fmt.Printf("  - %s: %s\n", e.Code, e.Message)
		}
	}
	fmt.Println("data:")
	dataJSON, _ := json.MarshalIndent(payload.Data, "  ", "  ")
	fmt.Println("  " + strings.ReplaceAll(string(dataJSON), "\n", "\n  "))
	if raw.IsError {
		fmt.Println("mcp_is_error: true")
	}
}

func hasErrorCode(payload contractPayload, code string) bool {
	for _, e := range payload.Errors {
		if e.Code == code {
			return true
		}
	}
	return false
}

func runSetup(args []string) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	serverURL := fs.String("server-url", "", "MCP HTTP endpoint URL")
	serverCmd := fs.String("server-cmd", "", "MCP stdio server command")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*serverURL) == "" && strings.TrimSpace(*serverCmd) == "" {
		return errors.New("one of --server-url or --server-cmd is required")
	}
	if strings.TrimSpace(*serverURL) != "" && strings.TrimSpace(*serverCmd) != "" {
		return errors.New("set only one: --server-url or --server-cmd")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "rw")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	cfgPath := filepath.Join(dir, "config.env")
	line := ""
	if strings.TrimSpace(*serverURL) != "" {
		line = "RW_MCP_SERVER_URL=\"" + strings.TrimSpace(*serverURL) + "\"\n"
	}
	if strings.TrimSpace(*serverCmd) != "" {
		line = "RW_MCP_SERVER_CMD=\"" + strings.TrimSpace(*serverCmd) + "\"\n"
	}
	if err := os.WriteFile(cfgPath, []byte("# rw CLI config\n"+line), 0o644); err != nil {
		return err
	}
	fmt.Printf("saved config: %s\n", cfgPath)
	return nil
}

func runDoctor(opts globalOpts) error {
	fmt.Println("rw doctor")
	if opts.serverURL != "" {
		fmt.Printf("transport: http\nserver_url: %s\n", opts.serverURL)
		return nil
	}
	if opts.serverCmd != "" {
		fmt.Printf("transport: stdio\nserver_cmd: %s\n", opts.serverCmd)
		return nil
	}
	return errors.New("no server configured; run: rw setup --server-url <url> OR rw setup --server-cmd <cmd>")
}

func printHelp() {
	fmt.Println(`rw - Go CLI for Research Warehouse MCP

Usage:
  rw setup --server-url <url>
  rw setup --server-cmd "python /path/to/server.py"
  rw doctor
  rw <command> [options]

Global flags:
  --server-url <url>  MCP HTTP endpoint (or RW_MCP_SERVER_URL)
  --server-cmd <cmd>  MCP stdio command (or RW_MCP_SERVER_CMD)
  --timeout <sec>     Request timeout seconds (default: 30)
  --retries <n>       Retries on INTERNAL_ERROR (default: 1)
  --json              Print contract payload as JSON only

Commands:
  setup
  doctor
  tools
  call --tool <name> --args '{"k":"v"}'
  company --ticker NVDA
  chain --entity NVDA [--direction both --max-depth 3]
  bottleneck --domain memory
  macro [--days 30]
  search [--sector photonics --tier chokepoint]
  semantic --query "CoWoS capacity" [--limit 5]
  earnings --ticker NVDA [--quarter Q4-2025]
  reports [--ticker NVDA | --topic CoWoS] [--weeks 8 --limit 5]
  options --ticker NVDA [--days 30]
  etf [--ticker SMH --days 30]
  technicals --ticker NVDA [--days 7]`)
}

func exitIfErr(err error) {
	if err != nil {
		exitErr(err)
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
