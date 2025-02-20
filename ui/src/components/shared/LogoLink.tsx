import { Link } from "react-router-dom";

import IconLogo from "src/assets/optimism.svg?react";
import { ROOT_PATH } from "src/utils/constants";

export function LogoLink() {
  return (
    <Link to={ROOT_PATH}>
      <IconLogo />
    </Link>
  );
}
