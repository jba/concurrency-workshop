package patterns

// title Pipelines

////////////////////////////////////
// heading Concept

// text
// You have many items.

// Each requires processing which can be done in parts.

// So:
// - Process the items concurrently.
// - *Process the parts concurrently.*
//   - Usually with channels.

////////////////////////////////////
// heading Example: validating files
