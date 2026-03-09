import type { RoleInfo } from "../types/game";

interface Props {
  roleInfo: RoleInfo;
}

export default function RoleCard({ roleInfo }: Props) {
  const isSpy = roleInfo.role === "spy";

  return (
    <div className={`role-card ${isSpy ? "role-spy" : "role-regular"}`}>
      {isSpy ? (
        <>
          <div className="role-icon">&#128373;&#65039;</div>
          <h2>Вы — шпион</h2>
          <p>Определите локацию!</p>
        </>
      ) : (
        <>
          <div className="role-icon">&#128205;</div>
          <h2>{roleInfo.location}</h2>
          <p>Вы — обычный игрок</p>
        </>
      )}
    </div>
  );
}
