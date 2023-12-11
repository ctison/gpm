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
      bin_dir: directories::UserDirs::new()
        .unwrap()
        .home_dir()
        .join(".local/bin"),
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

  let release_dir = directories::UserDirs::new().unwrap().home_dir().join(
    [".cache/gpm", &opts.repo, &release.tag_name]
      .iter()
      .collect::<PathBuf>(),
  );
  let download_dir = release_dir.join("download");
  std::fs::create_dir_all(&download_dir)?;
  let asset_path = download_dir.join(&assets[0].name);

  info!("Downloading {} asset to {asset_path:?}", &assets[0].name);
  let client = reqwest::Client::new();
  let mut request = client.get(assets[0].browser_download_url.clone());
  let mut file = std::fs::OpenOptions::new()
    .append(true)
    .create(true)
    .open(&asset_path)?;
  if let Ok(metadata) = file.metadata() {
    if metadata.len() > 0 {
      request = request.header("Range", format!("bytes={}-", metadata.len()));
    }
  }
  let mut response = request.send().await?;
  let mut i = 0;
  while let Some(chunk) = response.chunk().await? {
    i += 1;
    file.write_all(&chunk)?;
  }
  debug!("Wrote {} chunks", &i);
  if let Ok(metadata) = file.metadata() {
    if metadata.len() != assets[0].size as u64 {
      return Err(anyhow!(
        "Downloaded asset size ({}) does not match expected size ({})",
        metadata.len(),
        assets[0].size
      ));
    }
  }
  info!("Saved asset to {asset_path:?}");

  let extract_path = release_dir.join("extract");

  std::fs::create_dir_all(&opts.bin_dir)?;

  // TODO: Extract & Link
  let drop_stdout = gag::Gag::stdout().unwrap();
  if decompress::can_decompress(&asset_path) {
    std::fs::create_dir_all(&extract_path)?;
    drop(drop_stdout);
    info!("Decompressing {asset_path:?}");
    let drop_stdout = gag::Gag::stdout().unwrap();
    let files = decompress::list(
      &asset_path,
      &decompress::ExtractOptsBuilder::default().build()?,
    )?;
    let mut decompressed_files = Vec::with_capacity(files.entries.len());

    for file in files.entries {
      match decompress::decompress(
        &asset_path,
        &extract_path,
        &decompress::ExtractOptsBuilder::default()
          .filter(move |path| path == Path::new(&file))
          .build()?,
      ) {
        Err(err) => {
          if let decompress::DecompressError::IO(err) = err {
            if err.kind() != std::io::ErrorKind::AlreadyExists {
              return Err(err.into());
            }
          }
        }
        Ok(decompressed_file) => {
          decompressed_files.push(decompressed_file.files[0].clone());
        }
      };
    }

    drop(drop_stdout);

    for file in decompressed_files {
      debug!("Decompressed file: {:?}", file);
      link_binary(&file, opts.bin_dir.as_path())?;
    }
  } else {
    link_binary(&asset_path, &opts.bin_dir)?;
  }

  Ok(())
}

static FILTER_PATTERNS: &[&str] = &[
  ".txt",
  "md",
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

fn link_binary<P: AsRef<Path>, Q: AsRef<Path>>(file: P, bin_dir: Q) -> Result<()> {
  let file = file.as_ref();

  let mime = infer::get_from_path(file)?
    .ok_or_else(|| {
      anyhow!(
        "Could not infer mime type from file: {file:?}",
        file = file.display()
      )
    })?
    .mime_type();

  debug!("Mime type of {file:?}: {mime}");

  #[cfg(target_family = "windows")]
  if mime != "application/vnd.microsoft.portable-executable" {
    return Ok(());
  }

  #[cfg(target_os = "macos")]
  if mime != "application/x-mach-binary" {
    return Ok(());
  }

  #[cfg(all(target_family = "unix", not(target_os = "macos")))]
  if mime != "application/x-executable" {
    return Ok(());
  }

  let link_path = bin_dir.as_ref().join(
    Path::new(file)
      .file_name()
      .with_context(|| format!("Could not get file name from path: {file:?}"))?,
  );

  #[cfg(target_family = "windows")]
  let symlink_result = std::os::windows::fs::symlink_file(&file, &link_path)?;

  #[cfg(target_family = "unix")]
  let symlink_result = std::os::unix::fs::symlink(&file, &link_path);

  if let Err(err) = symlink_result {
    if err.kind() != std::io::ErrorKind::AlreadyExists {
      return Err(err.into());
    }
  }

  info!("Created symbolic link: {:?}", link_path);
  Ok(())
}

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
