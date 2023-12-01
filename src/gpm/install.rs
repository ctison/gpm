use crate::prelude::*;

#[derive(Debug, Clone, clap::Args)]
pub struct GpmInstallOpts {
  /// The repository to install from (eg. starship/starship)
  #[arg(name = "OWNER/REPO")]
  pub repo: String,

  /// The release tag to install
  #[arg(value_name = "String", long, default_value_t = GpmInstallOpts::default().tag)]
  pub tag: String,

  /// The architecture to filter by
  #[arg(value_name = "String", long, default_value_t = GpmInstallOpts::default().arch)]
  pub arch: String,

  /// The operating system to filter by
  #[arg(value_name = "String", long, default_value_t = GpmInstallOpts::default().os)]
  pub os: String,

  /// The libc to filter by.
  #[arg(value_name = "Enum", long, default_value = "gnu")]
  pub libc: Libc,

  /// The path to the bin directory.
  #[arg(value_name = "PathBuf", long, default_value_os_t = GpmInstallOpts::default().bin_dir)]
  pub bin_dir: PathBuf,
}

impl GpmInstallOpts {
  pub async fn install(&self) -> Result<()> {
    install(Some(self.clone())).await
  }
}

impl Default for GpmInstallOpts {
  fn default() -> Self {
    Self {
      repo: Default::default(),
      tag: "latest".into(),
      arch: std::env::consts::ARCH.to_string(),
      os: std::env::consts::OS.to_string(),
      libc: Default::default(),
      bin_dir: "~/local/bin".into(),
    }
  }
}

#[derive(Debug, Default, clap::ValueEnum, Clone)]
pub enum Libc {
  #[default]
  Gnu,
  Musl,
}

pub async fn install(opts: Option<GpmInstallOpts>) -> Result<()> {
  let opts = opts.unwrap_or_default();
  let arch = &opts.arch;
  let os = &opts.os;

  debug!("Architecture {arch}");
  debug!("Operating system {os}");

  let filter_arch = match arch.as_str() {
    "x86_64" => vec!["x86_64", "amd64", "x64"],
    "aarch64" => vec!["aarch64", "arm64"],
    "armv6l" | "armv7l" => vec![arch.as_str(), "armhf"],
    _ => vec![arch.as_str()],
  };

  let filter_os = match os.as_str() {
    "Linux" => vec!["linux"],
    "macos" => vec!["darwin", "apple", "macos"],
    "Windows" => vec!["windows", "win"],
    _ => vec![os.as_str()],
  };

  debug!("Filter architecture {filter_arch:?}");
  debug!("Filter operating system {filter_os:?}");

  let gh = octocrab::instance();

  let (owner, repo) = opts.repo.split_once('/').unwrap();
  let repo = gh.repos(owner, repo);
  let releases = repo.releases();

  // TODO: Handle pre-releases and interactive
  let release = if opts.tag == "latest" {
    releases.get_latest().await?
  } else {
    releases.get_by_tag(&opts.tag).await?
  };

  info!(
    "Installing {} {}",
    &opts.repo,
    &release.name.as_deref().unwrap_or(&release.tag_name)
  );

  let assets = release
    .assets
    .iter()
    .filter(|asset| {
      let name = asset.name.to_lowercase();
      filter_arch.iter().any(|arch| name.contains(arch))
        && filter_os.iter().any(|os| name.contains(os))
        && !FILTER_PATTERNS
          .iter()
          .any(|pattern| name.ends_with(pattern))
        && match opts.libc {
          Libc::Gnu => !name.contains("musl"),
          Libc::Musl => !name.contains("gnu"),
        }
    })
    .collect::<Vec<_>>();

  if assets.len() == 0 {
    return Err(anyhow!("No assets found for {arch} {os}"));
  }

  info!(
    "Filtered assets: {:?}",
    assets.iter().map(|asset| &asset.name).collect::<Vec<_>>()
  );

  // TODO: Handle filtered assets > 1 & interactive
  if assets.len() > 1 {
    return Err(anyhow!("Multiple assets found for {arch} {os}"));
  }

  let release_dir = ["~/.cache/gpm", &opts.repo, &release.tag_name]
    .iter()
    .collect::<PathBuf>();
  let download_dir = release_dir.join("download");
  fs::create_dir_all(&download_dir).await?;
  let asset_path = download_dir.join(&assets[0].name);

  info!("Downloading {} asset to {asset_path:?}", &assets[0].name);
  // TODO: Download asset
  info!("Saved asset to {asset_path:?}");

  let extract_path = release_dir.join("extract");
  fs::create_dir_all(&extract_path).await?;

  fs::create_dir_all(&opts.bin_dir).await?;

  // TODO: Extract & Link
  if [".tar.gz", ".tar.xz", ".tgz"]
    .iter()
    .any(|filetype| assets[0].name.ends_with(filetype))
  {
    info!("Extracting tarball");
  } else if assets[0].name.ends_with(".zip") {
    info!("Extracting zip");
  } else {
    info!("Linking binary");
  }

  Ok(())
}

static FILTER_PATTERNS: &[&str] = &[
  ".txt",
  ".sha256",
  ".md5",
  ".shasum",
  ".sig",
  ".sha1",
  ".sha512",
  ".sha256sum",
  ".asc",
  ".json",
  ".yaml",
  ".sh",
  ".1",
  ".2",
  ".3",
  ".5",
  ".6",
  ".7",
  ".8",
  ".vsix",
];

#[cfg(test)]
mod tests {
  use super::*;

  #[tokio::test]
  async fn test_() -> Result<()> {
    let opts = GpmInstallOpts::default();
    println!("{:?}", opts);
    Ok(())
  }
}
