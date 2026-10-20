{ ... }:
let
  pname = "seanime-denshi";
in
{
  perSystem =
    {
      pkgs,
      config,
      ...
    }:
    {
      packages = {
        ${pname} = pkgs.callPackage ./${pname}.nix { };
        default = config.packages.${pname};
      };
    };
}
