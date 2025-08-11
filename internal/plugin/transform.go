package plugin

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
    "io"
    "strings"
)

// TransformPlugin performs data-plane transformations: format convert, masking, compression.
// Config:
//   json_to_xml: bool
//   xml_to_json: bool
//   mask_fields: ["password", "token"]
//   add_units: map[string]string (e.g., {"latency_ms": "ms"})
//   gzip_compress: bool
//   gzip_decompress: bool
//   grpc_status_map: map[string]int (grpc-status => http status)
//   add_grpc_metadata: map[string]string (key-value to inject as headers)
type TransformPlugin struct {
    jsonToXML       bool
    xmlToJSON       bool
    maskFields      []string
    addUnits        map[string]string
    gzipCompress    bool
    gzipDecompress  bool
    grpcStatusMap   map[string]int
    addGRPCMetadata map[string]string
}

func (p *TransformPlugin) Name() string { return "transform" }

func (p *TransformPlugin) Init(cfg map[string]any) error {
    if v, ok := cfg["json_to_xml"].(bool); ok { p.jsonToXML = v }
    if v, ok := cfg["xml_to_json"].(bool); ok { p.xmlToJSON = v }
    if v, ok := cfg["gzip_compress"].(bool); ok { p.gzipCompress = v }
    if v, ok := cfg["gzip_decompress"].(bool); ok { p.gzipDecompress = v }
    if m, ok := cfg["grpc_status_map"].(map[string]any); ok {
        p.grpcStatusMap = map[string]int{}
        for k, vi := range m { if i, ok := vi.(int); ok { p.grpcStatusMap[k] = i } }
    }
    if m, ok := cfg["add_grpc_metadata"].(map[string]any); ok {
        p.addGRPCMetadata = map[string]string{}
        for k, vi := range m { if s, ok := vi.(string); ok { p.addGRPCMetadata[k] = s } }
    }
    if arr, ok := cfg["mask_fields"].([]any); ok {
        for _, v := range arr { if s, ok := v.(string); ok { p.maskFields = append(p.maskFields, s) } }
    }
    if m, ok := cfg["add_units"].(map[string]any); ok {
        p.addUnits = map[string]string{}
        for k, vi := range m { if s, ok := vi.(string); ok { p.addUnits[k] = s } }
    }
    return nil
}

func (p *TransformPlugin) BeforeDispatch(ctx *RequestContext) (bool, error) {
    // inject gRPC metadata if configured
    if len(p.addGRPCMetadata) > 0 && isGRPCContentType(ctx.Request.Header.Get("Content-Type")) {
        for k, v := range p.addGRPCMetadata {
            ctx.Request.Header.Set(k, v)
        }
    }
    return false, nil
}

func (p *TransformPlugin) AfterDispatch(ctx *RequestContext) {
    if ctx.Response == nil { return }
    // Map gRPC status to HTTP if provided via trailers
    if len(p.grpcStatusMap) > 0 {
        if st := ctx.Response.Trailer.Get("grpc-status"); st != "" {
            if mapped, ok := p.grpcStatusMap[st]; ok { ctx.Response.StatusCode = mapped }
        }
    }
    // Body transforms
    ct := strings.ToLower(ctx.Response.Header.Get("Content-Type"))
    body := ctx.Response.Body
    // Optional gzip decompress first
    if p.gzipDecompress && strings.Contains(ct, "gzip") {
        if dec, err := gunzip(body); err == nil { body = dec; removeContentEncoding(ctx.Response.Header) }
    }
    // JSON masking and unit tagging (only for JSON)
    if strings.Contains(ct, "application/json") {
        body = maskJSONFields(body, p.maskFields)
        // Unit tagging could be represented via extra fields or headers; simplified omitted here
    }
    // Format conversions (simplified placeholders)
    if p.jsonToXML && strings.Contains(ct, "application/json") {
        // simple wrapper for demo
        body = []byte("<json>" + string(body) + "</json>")
        ctx.Response.Header.Set("Content-Type", "application/xml")
    }
    if p.xmlToJSON && strings.Contains(ct, "application/xml") {
        body = []byte("{" + "\"xml\":" + jsonString(string(body)) + "}")
        ctx.Response.Header.Set("Content-Type", "application/json")
    }
    // Optional gzip compress last
    if p.gzipCompress {
        if enc, err := gzipBytes(body); err == nil {
            body = enc
            ctx.Response.Header.Set("Content-Encoding", "gzip")
        }
    }
    ctx.Response.Body = body
}

func isGRPCContentType(ct string) bool { return strings.HasPrefix(ct, "application/grpc") }

func gzipBytes(b []byte) ([]byte, error) {
    var buf bytes.Buffer
    zw := gzip.NewWriter(&buf)
    if _, err := zw.Write(b); err != nil { return nil, err }
    if err := zw.Close(); err != nil { return nil, err }
    return buf.Bytes(), nil
}

func gunzip(b []byte) ([]byte, error) {
    zr, err := gzip.NewReader(bytes.NewReader(b))
    if err != nil { return nil, err }
    defer zr.Close()
    return io.ReadAll(zr)
}

func removeContentEncoding(h map[string][]string) {
    delete(h, "Content-Encoding")
}

func maskJSONFields(b []byte, fields []string) []byte {
    if len(fields) == 0 { return b }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil { return b }
    for _, f := range fields {
        if _, ok := m[f]; ok { m[f] = "***" }
    }
    nb, err := json.Marshal(m)
    if err != nil { return b }
    return nb
}

func jsonString(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}

func init() { Register("transform", func() Plugin { return &TransformPlugin{} }) }

