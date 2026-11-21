# Seanime Denshi for nix

The [seanime](https://github.com/5rahim/seanime) desktop AppImage
**Seanime Denshi**, packaged for nix.

## INSTALL

First add this flake to your inputs,
then add it to your `environment.systemPackages` or `home.packages`:

```nix flake.nix
## flake.nix
{
  description = "Hello World!";

  inputs = {
    ## ... your other inputs

    seanime-denshi = {
      #NOTE: If fork hasen't been merged yet, change it to:
      # url = "github:LibereCode/seanime.nix";
      url = "github:5rahim/seanime";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    ## ... your other inputs
  };

  ## ... outputs ...
}
```

Another file that is imported into your config:

```nix modules/nixos_or_hm/seanime-denshi.nix
{ inputs, pkgs, ... }:
let
  system = pkgs.stdenv.hostPlatform.system;
in
{
  config = {

    ## For NixOS
    environment.systemPackages = [
      inputs.seanime-denshi.packages.${system}.seanime-denshi
    ];

    ## For Home-manager
    home.packages = [
      inputs.seanime-denshi.packages.${system}.seanime-denshi
    ];

    ## (Do not do both in the same module)

  };
}
```
