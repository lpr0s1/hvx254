export default function handler(req, res) {
  // GET : afficher un message personnalisé via ?msg=
  if (req.method === "GET") {
    const msg = req.query.msg || "Aucun message fourni";
    return res.status(200).json({
      status: "OK",
      message: msg
    });
  }

  // POST : afficher un message envoyé dans le body
  if (req.method === "POST") {
    const { message } = req.body || {};
    return res.status(200).json({
      status: "OK",
      message: message || "Aucun message fourni"
    });
  }

  return res.status(405).json({ error: "Méthode non autorisée" });
}
