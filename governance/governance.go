package governance

import (
	"github.com/gin-gonic/gin"

	"github.com/barnbridge/internal-api/db"
)

type Governance struct {
	db *db.DB
}

func New(db *db.DB) *Governance {
	return &Governance{db: db}
}

func (g *Governance) SetRoutes(engine *gin.Engine) {
	governance := engine.Group("/api/governance")
	governance.GET("/proposals", g.AllProposalsHandler)
	governance.GET("/proposals/:proposalID", g.ProposalDetailsHandler)
	governance.GET("/proposals/:proposalID/votes", g.VotesHandler)
	governance.GET("/proposals/:proposalID/events", g.HandleProposalEvents)
	governance.GET("/overview", g.HandleOverview)
	governance.GET("/voters", g.HandleVoters)
	// governance.GET("/abrogation-proposals", a.AllAbrogationProposals)
	// governance.GET("/abrogation-proposals/:proposalID", a.AbrogationProposalDetailsHandler)
	// governance.GET("/abrogation-proposals/:proposalID/votes", a.AbrogationVotesHandler)
	// governance.GET("/treasury/transactions", a.handleTreasuryTxs)
	// governance.GET("/treasury/tokens", a.handleTreasuryTokens)
}
