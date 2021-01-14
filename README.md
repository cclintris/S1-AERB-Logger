S1 Logger Module
---

#### Struct

- Logger
  - const
	  - OPT_HAS_REPORT_CALLER LogOptions = 0x0001
	  - OPT_HAS_SHORT_CALLER  LogOptions = 0x0002
	  - FILE     string = "file"
	  - FUNCTION string = "func"
	  - RESOURCE string = "res"
	  - CATEGORY string = "cat"
  - func `SetResource(resource string) *Logger`
  
    Set the resource which generates logs until cleared. The resource name should lead with a character, in UPPER case, to represent type of resource which is  followed by the UUID of resource and seperated by colon ':'.
    e.g. D:112233445566-2020R12345678

    **Resource Types:**
    
    | Type | Descripption |
    | :--: | :----------- |
    | D    | Device. |
    | U    | User.   |
    
  - func `ClearResource() *Logger`
    
    Clear resource field.
  - func `SetCategory(category string) *Logger`
    
    Set the category logs until cleared.
  - func `ClearCategory() *Logger`
    
    Clear category field.
  - func `ClearAll() *Logger`
    
    Clear resource and category fields.

#### API

- func `New() *Logger`
  
    Generate a singleton logger and set default options. default options is (`OPT_HAS_REPORT_CALLER|OPT_HAS_SHORT_CALLER`)
- func `NewWithOptions(options LogOptions) *Logger`
  
    Generate a singleton logger and set custom options.
