let lastMessage = "Aucun message pour le moment";

export const config = {
  api: { bodyParser: true } // On réactive le body parser pour lire du JSON
};

export default async function handler(req, res) {
  // GET = récupérer le dernier message
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage
    });
  }

  // POST = envoyer un message
  if (req.method === "POST") {
    try {
      const { message } = req.body;

      if (!message || typeof message !== "string") {
        return res.status(400).json({ error: "Message invalide" });
      }

      lastMessage = message;

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage
      });
    } catch (e) {
      return res.status(500).json({ error: "Erreur interne", details: e.message });
    }
  }

  return res.status(405).json({ error: "Méthode non autorisée" });
}
