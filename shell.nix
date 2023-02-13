{ pkgs ? import <nixpkgs> {} }:

let fetchPatchFromGitHub = { owner, repo, rev, sha256 }:
		pkgs.fetchpatch {
			url = "https://github.com/${owner}/${repo}/commit/${rev}.patch";
			inherit sha256;
		};

	goapi-gen = pkgs.buildGoModule {
		name = "goapi-gen";
		version = "081d60b";

		src = pkgs.fetchFromGitHub {
			owner  = "discord-gophers";
			repo   = "goapi-gen";
			rev    = "074e08fcdf14d45fc9b8812fa97c3c05f5f97626";
			sha256 = "109q1vxp8y131n6b0dxvg2zb9vqylxgn6q9n8y2d4v5yhh88436r";
		};

		vendorSha256 = "1dknfg3w97421c8dnld5kvx0psicvmxr7wzkhqipaxplcg3cqrr9";
	};

	sqlc = pkgs.buildGoModule {
		name = "sqlc";
		version = "1.12.0";

		src = pkgs.fetchFromGitHub {
			owner  = "kyleconroy";
			repo   = "sqlc";
			rev    = "45bd150";
			sha256 = "1np2xd9q0aaqfbcv3zcxjrfd1im9xr22g2jz5whywdr1m67a8lv2";
		};

		proxyVendor = true;
		vendorSha256 = "0mk1bs8ppis7dr3mcg73j8abvi1qpbg06adx8sxpzrfag4i9vg8k";
	};

	goose = pkgs.buildGoModule {
		name = "goose";
		version = "3.5.3";

		src = pkgs.fetchFromGitHub {
			owner  = "pressly";
			repo   = "goose";
			rev    = "5f1f43cfb2ba11d901b1ea2f28c88bf2577985cb";
			sha256 = "13hcbn4v78142brqjcjg1q297p4hs28n25y1fkb9i25l5k2bwk7f";
		};

		vendorSha256 = "1yng6dlmr4j8cq2f43jg5nvcaaik4n51y79p5zmqwdzzmpl8jgrv";
		subPackages = [ "cmd/goose" ];
	};

	nixos-shell = pkgs.buildGoModule {
		name = "nixos-shell";
		src = pkgs.fetchFromGitHub {
			owner  = "diamondburned";
			repo   = "nixos-shell";
			rev    = "e238cb522f7168fbc997101d00e6e2cc0d3e2ff9";
			sha256 = "02wqbfmc0c7q3896x6k2hxwcf1x202qfw0almb6rchlh7cqkva0w";
		};
		vendorSha256 = "0gjj1zn29vyx704y91g77zrs770y2rakksnn9dhg8r6na94njh5a";
	};
	
	jtd-codegen = pkgs.rustPlatform.buildRustPackage {
		pname = "jtd-codegen";
		version = "dev";

		# TODO: push the PR and replace this.
		src = pkgs.fetchFromGitHub {
			owner  = "jsontypedef";
			repo   = "json-typedef-codegen";
			rev    = "v0.4.1";
			sha256 = "1922k67diwrbcm6rq18pzr9627xzkv00k3y2dc4843hn25kqqha5";
		};

		patches = [
			(fetchPatchFromGitHub {
				owner  = "diamondburned";
				repo   = "json-typedef-codegen";
				rev    = "dcbba615d8a0a398eef6675670dbe0ea7d3e3a8e";
				sha256 = "1zb5fd4d9d5lxq7shcj0vkw4bqxw308pyhf25d6y67rgnkslmgx4";
			})
			(fetchPatchFromGitHub {
				owner  = "diamondburned";
				repo   = "json-typedef-codegen";
				rev    = "63148e9e727f1e0110d554d4eb031a734bdb60e1";
				sha256 = "16gjdz470j8zf5xlqbvs1pkvq0kpw065pbbb4jdh5yw450f9ywak";
			})
		];

		cargoHash = "sha256:1hax2whf0kfb2sw2a6rams2c46qk3360wkdxp0rgwfyjsxz5znk3";

		# These tests need Docker for some stupid reason.
		doCheck = false;
	};

	jsonnet-language-server =
		let src = pkgs.fetchFromGitHub {
			owner  = "grafana";
			repo   = "jsonnet-language-server";
			rev    = "v0.11.0";
			sha256 = "1gh5j9gn23f7az4hqiq2ibb41bm2w76vda2ls8bavh7qbfvjvwm0";
		};
		in pkgs.callPackage "${src}/nix" { };

in pkgs.mkShell {
	buildInputs = with pkgs; [
		go_1_18
		gopls
		goapi-gen
		goose
		sqlc
		pgformatter
		nixos-shell # for local PostgreSQL server
		nodejs
		yq-go
		yajsv
		jsonnet
		jtd-codegen
		jsonnet-language-server
	];

	shellHook = ''
		PATH="$PWD/node_modules/.bin:$PATH"
	'';
}
