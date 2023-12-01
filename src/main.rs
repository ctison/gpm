mod prelude;
use prelude::*;

use gpm::cli::Cli;

#[tokio::main]
async fn main() -> Result<()> {
  let cli = Cli::init()?;
  cli.command.run().await?;

  Ok(())
}
