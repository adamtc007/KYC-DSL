# Blocking Issues - Dictionary & DocMaster Services

**Date**: 2025-10-31  
**Status**: 🚨 **CRITICAL - gRPC calls blocking/hanging**  
**Severity**: HIGH - Services register but don't respond to RPC calls

---

## 🔴 Issue Summary

The newly added Dictionary and DocMaster services are **partially implemented** but have a critical blocking issue:

- ✅ Proto files created and generated
- ✅ Service implementations written
- ✅ Services register with gRPC server
- ✅ Services appear in reflection (`grpcurl list`)
- ❌ **RPC calls hang/block indefinitely**
- ❌ **No log output from service methods**

---

## 🧪 Reproduction Steps

### Current State
```bash
# Services are registered
grpcurl -plaintext [::1]:50070 list
# Output shows: kyc.dictionary.DictionaryService, kyc.docmaster.DocMasterService

# But RPC calls hang
grpcurl -plaintext -max-time 3 -d '{}' [::1]:50070 kyc.data.CaseService/ListAllCases
# Result: Timeout after 3 seconds, no response

grpcurl -plaintext -d '{"id":"attr-name"}' [::1]:50070 kyc.dictionary.DictionaryService/GetAttribute
# Result: Hangs indefinitely
```

### What Works
- ✅ Service reflection (list services)
- ✅ PostgreSQL connection
- ✅ gRPC server initialization
- ✅ Service registration

### What Doesn't Work
- ❌ ANY RPC call to ANY service (even existing ones like CaseService)
- ❌ No log output when RPC is called
- ❌ Timeout with max-time flag
- ❌ No errors thrown

---

## 🔍 Investigation Findings

### 1. Multiple Dataserver Processes
**Found**: 3 old dataserver processes still listening on port 50070
**Fixed**: Killed all with `pkill -9 dataserver`
**Status**: ✅ Resolved

### 2. Service Initialization
**Status**: ✅ All services initialize successfully
```
✅ Connected to PostgreSQL
✅ Data Service initialized successfully
🌐 gRPC server listening on :50070
```

### 3. Service Methods Present
**Status**: ✅ All methods are implemented
```
DataService.SaveCaseVersion ✅
DataService.GetCaseVersion ✅
DataService.ListCaseVersions ✅
DataService.ListAllCases ✅
```

### 4. Proto Interfaces
**Status**: ✅ Correctly aligned
```
CaseServiceServer interface has all methods
DataService implements CaseServiceServer
UnimplementedServers properly embedded
```

### 5. Port Binding
**Status**: ✅ Listening on correct port
```
TCP *:50070 (LISTEN) - IPv6
```

### 6. Database Connection
**Status**: ✅ Connected
```
📊 Connection pool: max=20, min=5
```

### 7. RPC Call Logging
**Status**: ❌ **CRITICAL** - No logs when RPC called
```
Expected: "📦 ListAllCases: limit=X, offset=Y, status_filter=Z"
Actual: (no output, hangs)
```

---

## 🎯 Root Cause Analysis

### Hypothesis 1: Goroutine Deadlock
- **Likelihood**: HIGH
- **Evidence**: RPC calls completely block with no output
- **Theory**: Possible mutex deadlock in service initialization or RPC handler
- **Test Needed**: Add pprof profiling to see goroutine stacks

### Hypothesis 2: Context Cancellation
- **Likelihood**: MEDIUM
- **Evidence**: No logs = handler not reached
- **Theory**: Context might be cancelled during initialization
- **Test Needed**: Add context timeout debugging

### Hypothesis 3: gRPC Reflection Issue
- **Likelihood**: LOW
- **Evidence**: Reflection works (can list services)
- **Theory**: Reflection works but actual RPC routing doesn't
- **Test Needed**: Compare working vs non-working services

### Hypothesis 4: Proto Mismatch
- **Likelihood**: LOW
- **Evidence**: All methods present, compilation succeeds
- **Theory**: Proto definitions don't match implementation
- **Test Needed**: Verify all proto fields match Go struct fields

---

## 📋 Files Involved

### Proto Definitions
- `proto_shared/dictionary_service.proto` - Created 17:33
- `proto_shared/docmaster_service.proto` - Created 17:33
- Generated: `api/pb/kycdictionary/*.pb.go`
- Generated: `api/pb/kycdocmaster/*.pb.go`

### Service Implementations
- `internal/dictionary/server.go` - 333 lines, looks correct
- `internal/docmaster/server.go` - 210 lines, looks correct

### Server Integration
- `cmd/dataserver/main.go` - Services commented out for debugging

---

## 🛠️ Debugging Checklist

### Phase 1: Isolate the Problem
- [ ] Add pprof profiling to dataserver
- [ ] Check goroutine stacks when RPC hangs
- [ ] Add debug logging to gRPC interceptors
- [ ] Test with simpler RPC (no DB access)
- [ ] Compare with working Rust service on port 50060

### Phase 2: Test Hypotheses
- [ ] Enable only Dictionary service (no DocMaster)
- [ ] Enable only DocMaster service (no Dictionary)
- [ ] Enable only existing services (remove new ones)
- [ ] Test with mock server (no database)
- [ ] Test with different proto message types

### Phase 3: Fix & Verify
- [ ] Remove mutex if deadlock found
- [ ] Fix context handling if needed
- [ ] Regenerate proto if mismatch found
- [ ] Add comprehensive logging
- [ ] Test all RPC calls

---

## 🔧 Current Workaround

**Temporarily Disabled**: Both new services are commented out in `cmd/dataserver/main.go`

```go
// TODO: Dictionary and DocMaster services temporarily disabled for debugging
// dictionaryService := dictionary.NewServer()
// pbDictionary.RegisterDictionaryServiceServer(grpcServer, dictionaryService)
// 
// docMasterService := docmaster.NewServer()
// pbDocMaster.RegisterDocMasterServiceServer(grpcServer, docMasterService)
```

**Status**: Even with services disabled, existing CaseService also hangs!
**Implication**: Problem is NOT with new services, but with ALL services now

---

## 🚨 Critical Discovery

**ALL RPC calls are now blocking**, including existing ones that worked before:
- ❌ `kyc.data.CaseService/ListAllCases` - HANGS
- ❌ `kyc.data.DictionaryService/ListAttributes` - HANGS (old service)
- ✅ `grpcurl list` - WORKS (reflection only)

**This suggests**:
1. Something changed in the server initialization
2. A deadlock or blocking call during startup
3. Proto generation issue affecting all services
4. Database connection pool issue

---

## 📝 Next Steps

1. **Add Pprof Debugging**
   - Import _ "net/http/pprof"
   - Start debug HTTP server on :6060
   - Check goroutine stacks at runtime

2. **Simplify Test Case**
   - Create minimal service with one RPC
   - No database access
   - No mutex locks
   - See if that works

3. **Check Recent Changes**
   - Did proto generation affect all proto files?
   - Did go.mod get corrupted?
   - Are there import conflicts?

4. **Compare with Working Example**
   - The Rust service on port 50060 works fine
   - Check what's different in gRPC setup
   - Might be a Go gRPC versioning issue

5. **Rollback Test**
   - Restore original cmd/dataserver/main.go (before new services)
   - Rebuild and test
   - See if old code still hangs

---

## 📊 Timeline

| Time | Event |
|------|-------|
| 17:33 | Created proto files and generated code |
| 17:35 | Created service implementations |
| 17:36 | Added services to dataserver, services started successfully |
| 17:53 | Discovered RPC calls hanging |
| 18:00 | Confirmed ALL RPC calls hang (not just new services) |
| 18:01 | Disabled new services for debugging |

---

## 🎯 Success Criteria for Fix

✅ Fix complete when:
1. `grpcurl list` returns service list (ALREADY WORKS)
2. `grpcurl ... kyc.data.CaseService/ListAllCases` returns results
3. `grpcurl ... kyc.dictionary.DictionaryService/GetAttribute` returns attribute
4. `grpcurl ... kyc.docmaster.DocMasterService/GetDocument` returns document
5. All RPC calls include log output (proving handler was called)
6. No timeouts or hangs
7. All existing tests still pass

---

## 📞 For Next Session

**Start with**: Investigating why ALL RPC calls hang, not just the new services

**Key Question**: What changed between the last working state and now?

**Recommendation**: 
1. Check git diff for dataserver changes
2. Add pprof to see what goroutine is blocking
3. Test with a minimal proto-based gRPC service
4. Compare with Rust service (port 50060) which works fine

---

**Status**: 🚨 BLOCKED - Needs debugging  
**Priority**: CRITICAL  
**Owner**: Next Session  
**Estimated Time to Debug**: 1-2 hours