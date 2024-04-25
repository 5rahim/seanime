<p align="center">
<img src="images/logo.png" alt="preview" width="75px"/>
</p>

<h2 align="center"><b>Seanime Web</b></h2>

<h4 align="center">Web interface for <a href="https://github.com/5rahim/seanime/">Seanime Server</a></h4>

```txt
app/(main)
├── 📁 _atoms			
├── 📁 _containers						<- shared smart components
├── 📁 _hooks							<- top-level queries and global state hooks
├── 📁 _features						<- shared dumb components
└── 📁 {route}
    ├── 📁 _containers					<- route specific containers
    │   ├── 📁 {container1}				<- group of smart components (e.g. list of items)
    │   │   ├── 📁 _components	
    │	│   └── 📄 {container1}.tsx	
    │   └── 📄 {container2}.tsx			<- standalone smart component (e.g. form)
    ├── 📁 _components					<- route specific dumb components (e.g. card)			
    ├── 📁 _lib							<- route specific utility functions / hooks / states
    └── 📄 page.tsx
```

- **_atoms**: where global states (jotai atoms) are defined.
- **_containers**: contains "smart" components or groups of components.
  - "Smart" components are components that interact with local/global state, fetch data or do mutations.
  - Related groups of components should be placed in the same folder and standalone components should be placed in the root.
- **_hooks**: top-level queries and global state hooks.
- **_features**: reusable "dumb" components that are used across multiple pages.
  - "Dumb" components are components that only receive props and render UI.
