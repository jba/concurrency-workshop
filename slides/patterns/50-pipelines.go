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
// !text

////////////////////////////////////
// heading Not a pipeline

// text Each goroutine does the entire task.

// image pipelines-1.png

////////////////////////////////////
// heading Using a pipeline

// text Each stage does one part.

// image pipelines-2.png

// html <br/>
// text &nbsp;

// text Use buffered channels to avoid stalls.

////////////////////////////////////
// heading Advantages of pipelines

// text
// - Better use of resources<br/>
//   Some goroutines do I/O, some compute
// - Plug-and-play architecture<br/>
//   Wire up separate components
// !text

////////////////////////////////////
// heading Example: OpenTelemetry Collector

// image pipelines-3.png

////////////////////////////////////
// heading Disadvantages of pipelines

// text
// - Lots of overhead<br/>
//   Each stage should involve a lot of work
//
// - More complex code<br/>
// !text
