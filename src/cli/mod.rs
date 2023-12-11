use crate::install::GpmInstallOpts;
use crate::prelude::*;
use clap::{builder::styling::AnsiColor, Parser};
use tracing_subscriber::layer::SubscriberExt;

impl Cli {
  pub fn init() -> Result<Self> {
    let _ = dotenvy::from_filename(".env.local");
    let _ = dotenvy::from_filename(".env");

    let cli = Cli::parse();

    let subscriber = tracing_subscriber::registry()
      .with(tracing_subscriber::fmt::layer().without_time().compact())
      .with(tracing_subscriber::EnvFilter::builder().parse(&cli.log_filter)?);

    tracing::subscriber::set_global_default(subscriber)
      .with_context(|| "failed to start tracing")?;

    Ok(cli)
  }
}

/// Install assets from a Github repository release
#[derive(Debug, clap::Parser)]
#[command(
  author,
  version,
  arg_required_else_help(true),
  styles(clap::builder::Styles::styled()
  .header(AnsiColor::Green.on_default())
  .usage(AnsiColor::Green.on_default())
  .literal(AnsiColor::Cyan.on_default())
  .placeholder(AnsiColor::Blue.bright(true).on_default()))
)]
pub struct Cli {
  /// Configure the log level.
  ///
  /// https://docs.rs/tracing-subscriber/latest/tracing_subscriber/filter/struct.EnvFilter.html
  #[arg(
    global = true,
    value_name = "String",
    short,
    long,
    env,
    default_value = "info"
  )]
  pub log_filter: String,

  #[command(subcommand)]
  pub command: Commands,
}

#[derive(Debug, clap::Subcommand)]
pub enum Commands {
  #[command(arg_required_else_help(true), visible_alias = "i")]
  Install {
    #[command(flatten)]
    opts: GpmInstallOpts,
  },
}

impl Commands {
  pub async fn run(&self) -> Result<()> {
    match self {
      Commands::Install { opts } => opts.install().await?,
    }
    Ok(())
  }
}
