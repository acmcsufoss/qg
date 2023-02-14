let overlay = self: super: 
	let nixpkgs_go_1_20 = import (super.fetchFromGitHub {
			owner  = "NixOS";
			repo   = "nixpkgs";
			rev    = "46c194bd83efc43da64cb137cab1178071f09c3b";
			sha256 = "01gjm68kr373fznbhgzbkblm1c7ry5mfvsdh4qpvy0h0wd7m8fsw";
		}) {};	
	in
	{
		go =
			if super.lib.versionAtLeast super.go.version "1.20"
			then super.go
			else self.go_1_20;
		go_1_20 =
			if super ? go_1_20
			then super.go_1_20
			else nixpkgs_go_1_20.go_1_20;
	};

in { pkgs ? import <nixpkgs> { overlays = [ overlay ]; } }:

let lib = pkgs.lib;

	fetchPatchFromGitHub = { owner, repo, rev, sha256 }:
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

	sqlc = pkgs.buildGoModule rec {
		name = "sqlc";
		version = "1.17.0";

		src = pkgs.fetchFromGitHub {
			owner  = "kyleconroy";
			repo   = "sqlc";
			rev    = "v" + version;
			sha256 = "0mmibx1ak3w8zsd14mkcjzr27zr4hmgdczk2s41w0wxy0d1yaxlj";
		};

		doCheck = false;
		proxyVendor = true;
		vendorSha256 = "0dq4mbccsrlli94yvrhxln33qql1psv8k4lbsrjbyyszl5fq6a0s";
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
	
	jtd-codegen = pkgs.rustPlatform.buildRustPackage rec {
		pname = "jtd-codegen-patched";
		version = "0.4.1";

		# TODO: push the PR and replace this.
		src = pkgs.fetchFromGitHub {
			owner  = "jsontypedef";
			repo   = "json-typedef-codegen";
			rev    = "v${version}";
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

		cargoHash = "sha256:0awsvzszca60mw7l48w23fmjll092gk7px77k4f88lxxdy63c1jp";
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
		go
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
