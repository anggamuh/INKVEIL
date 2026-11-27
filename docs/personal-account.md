# Account library

MangaDex account data uses a separate MySQL table created by initAccounts. Guest storage remains separate; no automatic import. Optimistic revision prevents overwriting concurrent updates. Settings store browser preferences and display runtime provider configuration, not server-side activation controls. Source/language defaults apply to plain homepage links. No background polling or image transport changes. Run go test ./... and go vet ./... before deployment.
