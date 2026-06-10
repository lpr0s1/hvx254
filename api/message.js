let lastMessage = "Aucun message pour le moment";

export default function handler(req, res) {
  // GET → renvoyer le dernier message reçu
  if (req.method === "GET") {
    return res.status(200).json({
      status: "OK",
      message: lastMessage
    });
  }

  // POST → recevoir un message
  if (req.method === "POST") {
    const { message } = req.body || {};

    if (!message) {
      return res.status(400).json({ error: "Aucun message reçu" });
    }

    lastMessage = message; // on sauvegarde le message

    return res.status(200).json({
      status: "OK",
      received: message
    });
  }

  return res.status(405).json({ error: "Méthode non autorisée" });
}
