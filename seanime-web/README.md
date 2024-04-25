<p align="center">
<img src="public/logo_2.png" alt="preview" width="150px"/>
</p>

<h2 align="center"><b>Seanime Web</b></h2>

<h4 align="center">Web interface</h4>

```txt
📁 api
    ├── 📁 client
    ├── 📁 generated
    └── 📁 hooks
📁 app/(main)	
    ├── 📁 _atoms
    ├── 📁 _features
    ├── 📁 _hooks
    ├── 📁 _listeners
    └── 📁 {route}
    	├── 📁 _containers
    	├── 📁 _components
    	├── 📁 _lib
    	└── 📄 page.tsx
📁 components
```

- `api`: API related code.
  - `client`: React-Query and Axios related code.
  - `generated`: Generated types and endpoints.
  - `hooks`: Data-fetching hooks.


- `app/_atoms`: Global Jotai atoms
  - Related groups of components should be placed in the same folder and standalone components should be placed in the root.
- `app/_hooks`: Top-level queries (loaders) and global state hooks.
- `app/_features`: Specialized components that are used across multiple pages.
- `app/_listeners`: Websocket listeners.


- `app/{route}/_components`: Route-specific components that only depend on props.
- `app/{route}/_containers`: Route-specific components that interact with global state and API.
- `app/{route}/_lib`: Route-specific utility functions, hooks, constants, and data-related functions.


- `components`: Primitive components, not tied to any feature or route.
