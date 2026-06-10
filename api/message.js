export default function handler(req, res) {
  if (req.method !== "POST") {
    return res.status(405).json({ error: "Méthode non autorisée" });
  }

  const { message } = req.body;

  if (!message) {
    return res.status(400).json({ error: "Aucun message reçu" });
  }

  console.log("Message reçu :", message);

  res.status(200).json({
    status: "OK",
    received: message
  });
}
